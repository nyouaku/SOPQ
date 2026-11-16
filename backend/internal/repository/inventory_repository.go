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
		for _, eff := range effects {
			if err := applyStockDeltaTx(tx, eff); err != nil {
				return err
			}
		}

		// 3. Tulis baris ledger (referensi ke dokumen ini).
		for _, entry := range ledgerEntries {
			entry.ReferenceID = &movement.ID
			entry.ReferenceCode = movement.Code
			if err := tx.Create(entry).Error; err != nil {
				return fmt.Errorf("failed to write ledger entry: %w", err)
			}
		}
		return nil
	})
}

// applyStockDeltaTx mengubah quantity_system satu (produk, gudang) dengan optimistic
// lock. Membuat baris stock bila belum ada (hanya untuk delta positif). Menolak bila
// hasil akhir negatif.
func applyStockDeltaTx(tx *gorm.DB, eff StockEffect) error {
	var stock domain.Stock
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id = ? AND warehouse_id = ?", eff.ProductID, eff.WarehouseID).
		First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if eff.Delta < 0 {
				return fmt.Errorf("stok %s di gudang tsb belum ada, tidak bisa dikurangi", eff.ProductID)
			}
			stock = domain.Stock{
				ProductID:      eff.ProductID,
				WarehouseID:    eff.WarehouseID,
				QuantitySystem: 0,
				Version:        0,
			}
			if err := tx.Create(&stock).Error; err != nil {
				return fmt.Errorf("failed to create stock row: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load stock: %w", err)
		}
	}

	newQty := stock.QuantitySystem + eff.Delta
	if newQty < 0 {
		return fmt.Errorf("stok tidak boleh negatif (stok %d, perubahan %+d)", stock.QuantitySystem, eff.Delta)
	}

	result := tx.Model(&domain.Stock{}).
		Where("id = ? AND version = ?", stock.ID, stock.Version).
		Updates(map[string]any{
			"quantity_system": newQty,
			"version":         stock.Version + 1,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update stock: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("stok berubah oleh proses lain, coba lagi")
	}
	return nil
}
