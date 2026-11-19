// ===== FILE: backend/internal/repository/inventory_repository.go =====
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InventoryFilter menyaring daftar dokumen pergerakan.
type InventoryFilter struct {
	WarehouseID  *uuid.UUID
	MovementType string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

// StockEffect adalah dampak satu item terhadap quantity_system satu gudang.
// Delta positif = stok naik (IN / tujuan transfer), negatif = turun (OUT / asal transfer).
type StockEffect struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Delta       int
}

type InventoryRepository interface {
	// Create menyimpan header + item dokumen pergerakan (tanpa efek stok).
	Create(ctx context.Context, movement *domain.StockMovement) error
	// FindByID mengembalikan header + item.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error)
	// List mengembalikan daftar dokumen pergerakan (tanpa item, paginated).
	List(ctx context.Context, f InventoryFilter) ([]domain.StockMovement, int64, error)
	// NextCode menghasilkan kode dokumen unik per hari (pola MV-YYYYMMDD-NNNN).
	NextCode(ctx context.Context, day time.Time) (string, error)
	// PostMovementTx menjalankan seluruh posting dalam SATU transaksi:
	// simpan header+item, ubah quantity_system tiap efek stok (optimistic lock),
	// tulis baris ledger dengan balance_after. Semua atau tidak sama sekali.
	PostMovementTx(ctx context.Context, movement *domain.StockMovement, effects []StockEffect, ledgerEntries []*domain.StockLedger) error
	// GridCells mengembalikan agregat isian per produk × tanggal untuk pre-fill grid.
	GridCells(ctx context.Context, warehouseID uuid.UUID, month time.Time, movementType string) ([]domain.GridCell, error)
	// Summary menghitung KPI dashboard inventory per gudang + bulan.
	Summary(ctx context.Context, warehouseID uuid.UUID, month time.Time) (*domain.InventorySummary, error)
	// CancelMovement membalik efek dokumen (stok & ledger) lalu menandainya
	// 'cancelled' — semuanya dalam satu transaksi. Dokumen tidak dihapus supaya
	// audit trail utuh; pembalikan dicatat sebagai baris ledger terpisah.
	CancelMovement(ctx context.Context, movementID uuid.UUID, userID uuid.UUID, reason string) error
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) Create(ctx context.Context, movement *domain.StockMovement) error {
	if err := r.db.WithContext(ctx).Create(movement).Error; err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}
	return nil
}

func (r *inventoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error) {
	var m domain.StockMovement
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Preload("Warehouse").
		Preload("ToWarehouse").
		Preload("CreatedByUser").
		Where("id = ?", id).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find movement: %w", err)
	}
	return &m, nil
}

func (r *inventoryRepository) List(ctx context.Context, f InventoryFilter) ([]domain.StockMovement, int64, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 25
	}
	q := r.db.WithContext(ctx).Model(&domain.StockMovement{})
	if f.WarehouseID != nil {
		q = q.Where("warehouse_id = ?", *f.WarehouseID)
	}
	if f.MovementType != "" {
		q = q.Where("movement_type = ?", f.MovementType)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count movements: %w", err)
	}
	var rows []domain.StockMovement
	if err := q.
		Preload("Warehouse").
		Order("created_at DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list movements: %w", err)
	}
	return rows, total, nil
}

// NextCode menghasilkan kode dokumen unik per hari (MV-YYYYMMDD-NNNN) secara atomik.
func (r *inventoryRepository) NextCode(ctx context.Context, day time.Time) (string, error) {
	dayStr := day.Format("2006-01-02")
	var seq struct{ LastValue int }
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO stock_movement_code_seq (day, last_value) VALUES (?, 1)
		ON CONFLICT (day) DO UPDATE
		SET last_value = stock_movement_code_seq.last_value + 1
		RETURNING last_value`, dayStr).Scan(&seq).Error
	if err != nil {
		return "", fmt.Errorf("failed to generate movement code: %w", err)
	}
	return fmt.Sprintf("MV-%s-%04d", day.Format("20060102"), seq.LastValue), nil
}

// PostMovementTx menjalankan seluruh posting dalam satu transaksi:
//   1. simpan header + item,
//   2. ubah quantity_system tiap StockEffect (optimistic lock; tolak bila hasil negatif),
//   3. tulis semua baris ledger (balance_after sudah dihitung pemanggil).
// Semua atau tidak sama sekali — gagal di mana pun membatalkan seluruh posting.
func (r *inventoryRepository) PostMovementTx(ctx context.Context, movement *domain.StockMovement, effects []StockEffect, ledgerEntries []*domain.StockLedger) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Header + item (item cascade ikut dibuat).
		if err := tx.Create(movement).Error; err != nil {
			return fmt.Errorf("failed to create movement: %w", err)
		}

		// 2. Terapkan efek stok per item dengan optimistic lock.
		//    Saldo baru dikembalikan agar baris ledger mencatat balance_after yang benar
		//    (kolomnya NOT NULL — tanpa ini kartu stok selalu 0).
		balances := make([]int, len(effects))
		for i, eff := range effects {
			balance, err := applyStockDeltaTx(tx, eff)
			if err != nil {
				return err
			}
			balances[i] = balance
		}

		// 3. Tulis baris ledger (referensi ke dokumen ini).
		//    effects dan ledgerEntries dibangun berdampingan, indeksnya sejajar.
		for i, entry := range ledgerEntries {
			entry.ReferenceID = &movement.ID
			entry.ReferenceCode = movement.Code
			if i < len(balances) {
				entry.BalanceAfter = balances[i]
			}
			if err := tx.Create(entry).Error; err != nil {
				return fmt.Errorf("failed to write ledger entry: %w", err)
			}
		}
		return nil
	})
}

// applyStockDeltaTx mengubah quantity_system satu (produk, gudang) dengan optimistic
// lock. Membuat baris stock bila belum ada (hanya untuk delta positif). Menolak bila
// hasil akhir negatif. Mengembalikan saldo setelah perubahan (untuk balance_after ledger).
func applyStockDeltaTx(tx *gorm.DB, eff StockEffect) (int, error) {
	var stock domain.Stock
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id = ? AND warehouse_id = ?", eff.ProductID, eff.WarehouseID).
		First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if eff.Delta < 0 {
				return 0, fmt.Errorf("stok %s di gudang tsb belum ada, tidak bisa dikurangi", eff.ProductID)
			}
			stock = domain.Stock{
				ProductID:      eff.ProductID,
				WarehouseID:    eff.WarehouseID,
				QuantitySystem: 0,
				Version:        0,
			}
			if err := tx.Create(&stock).Error; err != nil {
				return 0, fmt.Errorf("failed to create stock row: %w", err)
			}
		} else {
			return 0, fmt.Errorf("failed to load stock: %w", err)
		}
	}

	newQty := stock.QuantitySystem + eff.Delta
	if newQty < 0 {
		return 0, fmt.Errorf("stok tidak boleh negatif (stok %d, perubahan %+d)", stock.QuantitySystem, eff.Delta)
	}

	result := tx.Model(&domain.Stock{}).
		Where("id = ? AND version = ?", stock.ID, stock.Version).
		Updates(map[string]any{
			"quantity_system": newQty,
			"version":         stock.Version + 1,
		})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to update stock: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return 0, fmt.Errorf("stok berubah oleh proses lain, coba lagi")
	}
	return newQty, nil
}

// GridCells mengembalikan agregat isian per produk × tanggal untuk pre-fill grid
// pada bulan + gudang + tipe pergerakan tertentu. Sumber: stock_movement_item yang
// punya movement_date.
func (r *inventoryRepository) GridCells(ctx context.Context, warehouseID uuid.UUID, month time.Time, movementType string) ([]domain.GridCell, error) {
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var rows []domain.GridCell
	q := r.db.WithContext(ctx).
		Table("stock_movement_item AS smi").
		Select(`
			smi.product_id AS product_id,
			p.name AS name,
			p.product_code AS product_code,
			EXTRACT(DAY FROM smi.movement_date)::int AS day,
			SUM(CASE WHEN sm.movement_type = 'in' THEN smi.quantity ELSE 0 END)::int AS in_qty,
			SUM(CASE WHEN sm.movement_type = 'out' THEN smi.quantity ELSE 0 END)::int AS out_qty,
			MAX(smi.unit_cost) AS unit_cost`).
		Joins("JOIN stock_movement sm ON sm.id = smi.movement_id").
		Joins("JOIN products p ON p.id = smi.product_id").
		Where("sm.warehouse_id = ?", warehouseID).
		Where("sm.movement_type = ?", movementType).
		Where("smi.movement_date IS NOT NULL").
		Where("smi.movement_date >= ? AND smi.movement_date < ?", monthStart, monthEnd).
		Group("smi.product_id, p.name, p.product_code, day").
		Order("p.name ASC, day ASC")

	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to load grid cells: %w", err)
	}
	return rows, nil
}

// Summary menghitung KPI dashboard inventory: total IN/OUT/transfer + nilai + top keluar
// + stok menipis, untuk satu gudang dalam sebulan.
func (r *inventoryRepository) Summary(ctx context.Context, warehouseID uuid.UUID, month time.Time) (*domain.InventorySummary, error) {
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	summary := &domain.InventorySummary{}

	// Agregat per tipe pergerakan (header).
	var typeRows []struct {
		MovementType string
		Count        int64
		TotalIn      int
		TotalOut     int
		ValueIn      float64
		ValueOut     float64
	}
	err := r.db.WithContext(ctx).
		Table("stock_movement sm").
		Select(`
			sm.movement_type,
			COUNT(DISTINCT sm.id) AS count,
			COALESCE(SUM(smi.quantity) FILTER (WHERE sm.movement_type='in'),0)::int AS total_in,
			COALESCE(SUM(smi.quantity) FILTER (WHERE sm.movement_type='out'),0)::int AS total_out,
			COALESCE(SUM(smi.total_cost) FILTER (WHERE sm.movement_type='in'),0) AS value_in,
			COALESCE(SUM(smi.total_cost) FILTER (WHERE sm.movement_type='out'),0) AS value_out`).
		Joins("LEFT JOIN stock_movement_item smi ON smi.movement_id = sm.id").
		Where("sm.warehouse_id = ? AND sm.created_at >= ? AND sm.created_at < ? AND sm.deleted_at IS NULL",
			warehouseID, monthStart, monthEnd).
		Group("sm.movement_type").
		Scan(&typeRows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load summary: %w", err)
	}

	for _, tr := range typeRows {
		summary.MovementsCount += tr.Count
		switch tr.MovementType {
		case domain.MovementTypeIn:
			summary.TotalIn = tr.TotalIn
			summary.TotalValueIn = tr.ValueIn
		case domain.MovementTypeOut:
			summary.TotalOut = tr.TotalOut
			summary.TotalValueOut = tr.ValueOut
		case domain.MovementTypeTransfer:
			summary.TotalTransfer = tr.TotalOut // qty keluar dari gudang asal
		}
	}

	// Top 5 produk paling banyak keluar.
	var topOut []domain.TopMovement
	err = r.db.WithContext(ctx).
		Table("stock_movement_item AS smi").
		Select("smi.product_id AS product_id, p.name AS name, SUM(smi.quantity)::int AS quantity").
		Joins("JOIN stock_movement sm ON sm.id = smi.movement_id").
		Joins("JOIN products p ON p.id = smi.product_id").
		Where("sm.warehouse_id = ? AND sm.movement_type = 'out' AND sm.created_at >= ? AND sm.created_at < ?",
			warehouseID, monthStart, monthEnd).
		Group("smi.product_id, p.name").
		Order("quantity DESC").Limit(5).Scan(&topOut).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load top out: %w", err)
	}
	summary.TopOut = topOut

	// Stok menipis: quantity_system < minimum_stock (produk aktif).
	var low []domain.LowStockItem
	err = r.db.WithContext(ctx).
		Table("stock AS s").
		Select(`s.product_id AS product_id, p.name AS name,
			s.quantity_system AS quantity_system, p.minimum_stock AS minimum_stock`).
		Joins("JOIN products p ON p.id = s.product_id").
		Where("s.warehouse_id = ? AND s.quantity_system < p.minimum_stock AND p.deleted_at IS NULL",
			warehouseID).
		Order("s.quantity_system ASC").Limit(10).Scan(&low).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load low stock: %w", err)
	}
	summary.LowStock = low

	return summary, nil
}

// CancelMovement membalik seluruh efek dokumen lalu menandainya 'cancelled'.
// Dokumen TIDAK dihapus — audit trail wajib utuh. Pembalikan efek:
//   - in       : stok gudang turun sebesar qty
//   - out      : stok gudang naik sebesar qty
//   - transfer : gudang asal naik, gudang tujuan turun
// Tiap efek dibalik lewat applyStockDeltaTx, lalu ditulis satu baris ledger
// 'adjustment' yang mereferensikan dokumen asal, sehingga kartu stok menunjukkan
// pasangan masuk/keluar yang bisa ditelusuri.
func (r *inventoryRepository) CancelMovement(ctx context.Context, movementID, userID uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var mv domain.StockMovement
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			Where("id = ?", movementID).
			First(&mv).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return domain.ErrMovementNotFound
			}
			return fmt.Errorf("failed to load movement: %w", err)
		}
		switch mv.Status {
		case domain.MovementStatusCancelled:
			return domain.ErrMovementAlreadyCancelled
		case domain.MovementStatusPosted:
		default:
			return domain.ErrMovementNotCancellable
		}

		note := "koreksi " + mv.Code
		if reason != "" {
			note += ": " + reason
		}

		for _, it := range mv.Items {
			// Sisi-sisi yang harus dibalik: (gudang, arah ledger). Arah ledger adalah
			// kebalikan dari dokumen: IN asli -> qty masuk dibalik jadi keluar.
			type side struct {
				warehouseID uuid.UUID
				qtyIn       int
				qtyOut      int
			}
			var sides []side
			switch mv.MovementType {
			case domain.MovementTypeIn:
				sides = []side{{mv.WarehouseID, 0, it.Quantity}}
			case domain.MovementTypeOut:
				sides = []side{{mv.WarehouseID, it.Quantity, 0}}
			case domain.MovementTypeTransfer:
				sides = []side{{mv.WarehouseID, it.Quantity, 0}}
				if mv.ToWarehouseID != nil {
					sides = append(sides, side{*mv.ToWarehouseID, 0, it.Quantity})
				}
			default:
				return fmt.Errorf("tipe pergerakan tidak dikenal: %s", mv.MovementType)
			}

			for _, s := range sides {
				delta := s.qtyIn - s.qtyOut // positif = stok naik
				if _, err := applyStockDeltaTx(tx, StockEffect{
					ProductID:   it.ProductID,
					WarehouseID: s.warehouseID,
					Delta:       delta,
				}); err != nil {
					return fmt.Errorf("gagal membalik stok %s: %w", it.ProductID, err)
				}

				var balance int
				if err := tx.Model(&domain.Stock{}).
					Select("quantity_system").
					Where("product_id = ? AND warehouse_id = ?", it.ProductID, s.warehouseID).
					Scan(&balance).Error; err != nil {
					return fmt.Errorf("failed to read stock balance: %w", err)
				}

				entry := &domain.StockLedger{
					ProductID:       it.ProductID,
					WarehouseID:     s.warehouseID,
					TransactionType: domain.LedgerTypeAdjustment,
					ReferenceType:   "inventory_cancel",
					ReferenceID:     &mv.ID,
					ReferenceCode:   mv.Code,
					QuantityIn:      s.qtyIn,
					QuantityOut:     s.qtyOut,
					BalanceAfter:    balance,
					UnitCost:        it.UnitCost,
					TotalCost:       float64(it.Quantity) * it.UnitCost,
					Note:            note,
					CreatedBy:       &userID,
				}
				if err := tx.Create(entry).Error; err != nil {
					return fmt.Errorf("failed to write reversal ledger: %w", err)
				}
			}
		}

		update := map[string]any{
			"status":     domain.MovementStatusCancelled,
			"updated_at": time.Now(),
		}
		if reason != "" {
			if mv.Note != "" {
				update["note"] = mv.Note + "\n[dibatalkan] " + reason
			} else {
				update["note"] = "[dibatalkan] " + reason
			}
		}
		if err := tx.Model(&domain.StockMovement{}).
			Where("id = ?", mv.ID).
			Updates(update).Error; err != nil {
			return fmt.Errorf("failed to mark movement cancelled: %w", err)
		}
		return nil
	})
}
