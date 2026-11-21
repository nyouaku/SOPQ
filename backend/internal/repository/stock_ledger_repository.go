// ===== FILE: backend/internal/repository/stock_ledger_repository.go =====
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"gorm.io/gorm"
)

type StockLedgerRepository interface {
	// Create menyimpan satu entri ledger dalam transaksi.
	Create(ctx context.Context, entry *domain.StockLedger) error
	// CreateInTx menyimpan entri ledger dalam transaksi yang sudah berjalan (tx).
	CreateInTx(tx *gorm.DB, entry *domain.StockLedger) error
	// ListByProduct mengembalikan riwayat ledger satu produk di satu gudang.
	ListByProduct(ctx context.Context, productID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error)
	// ListByWarehouse mengembalikan riwayat ledger di satu gudang.
	ListByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error)
	// GetBalance mengembalikan saldo stok terakhir (balance_after terbaru) untuk produk di gudang.
	GetBalance(ctx context.Context, productID, warehouseID uuid.UUID) (int, error)
	// ListByDateRange mengembalikan riwayat ledger dalam rentang tanggal.
	ListByDateRange(ctx context.Context, warehouseID *uuid.UUID, from, to time.Time, limit, offset int) ([]domain.StockLedger, int64, error)
	// LedgerSummary menghitung saldo awal, total masuk, total keluar, dan saldo akhir
	// per produk di satu gudang untuk periode [from, to) — murni derived dari ledger.
	LedgerSummary(ctx context.Context, warehouseID uuid.UUID, from, to time.Time) ([]domain.LedgerSummaryRow, error)
}

type stockLedgerRepository struct {
	db *gorm.DB
}

func NewStockLedgerRepository(db *gorm.DB) StockLedgerRepository {
	return &stockLedgerRepository{db: db}
}

func (r *stockLedgerRepository) Create(ctx context.Context, entry *domain.StockLedger) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create stock ledger entry: %w", err)
	}
	return nil
}

func (r *stockLedgerRepository) CreateInTx(tx *gorm.DB, entry *domain.StockLedger) error {
	if err := tx.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create stock ledger entry (tx): %w", err)
	}
	return nil
}

func (r *stockLedgerRepository) ListByProduct(ctx context.Context, productID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var rows []domain.StockLedger
	var total int64
	q := r.db.WithContext(ctx).Model(&domain.StockLedger{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ledger: %w", err)
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list ledger: %w", err)
	}
	return rows, total, nil
}

func (r *stockLedgerRepository) ListByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var rows []domain.StockLedger
	var total int64
	q := r.db.WithContext(ctx).Model(&domain.StockLedger{}).
		Where("warehouse_id = ?", warehouseID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ledger: %w", err)
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list ledger: %w", err)
	}
	return rows, total, nil
}

func (r *stockLedgerRepository) GetBalance(ctx context.Context, productID, warehouseID uuid.UUID) (int, error) {
	var last domain.StockLedger
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Order("created_at DESC").First(&last).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get ledger balance: %w", err)
	}
	return last.BalanceAfter, nil
}

// ListByDateRange mengembalikan riwayat ledger dalam rentang tanggal.
// LedgerSummary menghitung saldo awal, total masuk/keluar, dan saldo akhir per produk
// untuk satu gudang pada periode [from, to). Semua dihitung dari stock_ledger —
// bukan stock live — supaya angka historis tidak berubah saat transaksi baru masuk.
func (r *stockLedgerRepository) LedgerSummary(ctx context.Context, warehouseID uuid.UUID, from, to time.Time) ([]domain.LedgerSummaryRow, error) {
	var rows []domain.LedgerSummaryRow
	err := r.db.WithContext(ctx).Raw(`
		WITH scope AS (
			SELECT sl.*
			FROM stock_ledger sl
			WHERE sl.warehouse_id = ? AND sl.created_at < ?
		),
		per_product AS (
			SELECT
				product_id,
				COALESCE(SUM(quantity_in - quantity_out) FILTER (WHERE created_at < ?), 0)::int AS opening,
				COALESCE(SUM(quantity_in) FILTER (WHERE created_at >= ? AND created_at < ?), 0)::int AS total_in,
				COALESCE(SUM(quantity_out) FILTER (WHERE created_at >= ? AND created_at < ?), 0)::int AS total_out,
				COALESCE(SUM(quantity_in * unit_cost) FILTER (WHERE created_at >= ? AND created_at < ?), 0)::numeric AS value_in,
				COALESCE(SUM(quantity_out * unit_cost) FILTER (WHERE created_at >= ? AND created_at < ?), 0)::numeric AS value_out
			FROM scope
			GROUP BY product_id
		)
		SELECT
			p.id AS product_id,
			p.product_code AS product_code,
			p.name AS product_name,
			p.unit AS unit,
			pp.opening AS opening,
			pp.total_in AS total_in,
			pp.total_out AS total_out,
			(pp.opening + pp.total_in - pp.total_out)::int AS closing,
			pp.value_in AS value_in,
			pp.value_out AS value_out
		FROM per_product pp
		JOIN products p ON p.id = pp.product_id
		WHERE p.deleted_at IS NULL
		ORDER BY p.name ASC`,
		warehouseID, to,
		from,
		from, to,
		from, to,
		from, to,
		from, to,
	).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load ledger summary: %w", err)
	}
	return rows, nil
}

func (r *stockLedgerRepository) ListByDateRange(ctx context.Context, warehouseID *uuid.UUID, from, to time.Time, limit, offset int) ([]domain.StockLedger, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var rows []domain.StockLedger
	var total int64
	q := r.db.WithContext(ctx).Model(&domain.StockLedger{}).
		Where("created_at >= ? AND created_at <= ?", from, to)
	if warehouseID != nil {
		q = q.Where("warehouse_id = ?", *warehouseID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ledger: %w", err)
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list ledger: %w", err)
	}
	return rows, total, nil
}
