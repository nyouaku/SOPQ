// ===== FILE: backend/internal/repository/stock_opname_repository.go =====
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockOpnameRepository interface {
	Create(ctx context.Context, stockOpname *domain.StockOpname) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.StockOpname, error)
	Update(ctx context.Context, stockOpname *domain.StockOpname) error
	List(ctx context.Context, filter StockOpnameFilter) ([]domain.StockOpname, int64, error)
	// NextCode mengambil nomor urut sesi berikutnya untuk hari tsb (atomik).
	NextCode(ctx context.Context, day time.Time) (string, error)
	// ApproveWithStock menyetujui sesi DAN menulis hasil hitung fisik ke tabel stock
	// dalam satu transaksi (transisi completed -> approved).
	ApproveWithStock(ctx context.Context, sessionID, approverID uuid.UUID) error
}

type StockOpnameFilter struct {
	WarehouseID *uuid.UUID
	Status      string
	StartedBy   *uuid.UUID
	Limit       int
	Offset      int
}

type stockOpnameRepository struct {
	db *gorm.DB
}

func NewStockOpnameRepository(db *gorm.DB) StockOpnameRepository {
	return &stockOpnameRepository{db: db}
}

func (r *stockOpnameRepository) Create(ctx context.Context, stockOpname *domain.StockOpname) error {
	if err := r.db.WithContext(ctx).Create(stockOpname).Error; err != nil {
		// 23505 = unique_violation. Hanya kode sesi yang punya constraint unik
		// di tabel ini, jadi pemanggil bisa retry dengan kode baru.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSessionCodeTaken
		}
		return fmt.Errorf("failed to create stock opname: %w", err)
	}
	return nil
}

func (r *stockOpnameRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockOpname, error) {
	var stockOpname domain.StockOpname
	if err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Location").
		Preload("StartedUser").
		Preload("ApprovedUser").
		Where("id = ?", id).
		First(&stockOpname).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find stock opname: %w", err)
	}
	return &stockOpname, nil
}

func (r *stockOpnameRepository) Update(ctx context.Context, stockOpname *domain.StockOpname) error {
	if err := r.db.WithContext(ctx).Save(stockOpname).Error; err != nil {
		return fmt.Errorf("failed to update stock opname: %w", err)
	}
	return nil
}

func (r *stockOpnameRepository) List(ctx context.Context, filter StockOpnameFilter) ([]domain.StockOpname, int64, error) {
	var stockOpnames []domain.StockOpname
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.StockOpname{})

	if filter.WarehouseID != nil {
		query = query.Where("warehouse_id = ?", filter.WarehouseID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.StartedBy != nil {
		query = query.Where("started_by = ?", filter.StartedBy)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count stock opnames: %w", err)
	}

	if err := query.
		Preload("Warehouse").
		Preload("Location").
		Preload("StartedUser").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order("created_at DESC").
		Find(&stockOpnames).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list stock opnames: %w", err)
	}

	return stockOpnames, total, nil
}

// NextCode menaikkan counter harian secara atomik dan mengembalikan kode sesi.
// ON CONFLICT ... RETURNING membuat satu round trip tanpa retry loop.
func (r *stockOpnameRepository) NextCode(ctx context.Context, day time.Time) (string, error) {
	dayStr := day.Format("2006-01-02")

	var seq struct{ LastValue int }
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO stock_opname_code_seq (day, last_value) VALUES (?, 1)
		ON CONFLICT (day) DO UPDATE
		SET last_value = stock_opname_code_seq.last_value + 1
		RETURNING last_value`, dayStr).Scan(&seq).Error
	if err != nil {
		return "", fmt.Errorf("failed to generate session code: %w", err)
	}

	return fmt.Sprintf("SO-%s-%04d", day.Format("20060102"), seq.LastValue), nil
}

// ApproveWithStock menyetujui sesi dan menulis quantity_system = hasil hitung fisik.
// Semua langkah dalam satu transaksi: kalau satu produk gagal, sesi tidak jadi approved.
func (r *stockOpnameRepository) ApproveWithStock(ctx context.Context, sessionID, approverID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session domain.StockOpname
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", sessionID).
			First(&session).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return domain.ErrSessionNotFound
			}
			return fmt.Errorf("failed to load session: %w", err)
		}

		if session.Status == "approved" {
			return domain.ErrSessionAlreadyApproved
		}
		if session.Status != "completed" {
			return domain.ErrSessionNotApprovable
		}

		type row struct {
			ProductID     uuid.UUID
			StockPhysical int
		}
		var rows []row
		if err := tx.Model(&domain.StockOpnameDetail{}).
			Select("product_id, stock_physical").
			Where("stock_opname_id = ?", sessionID).
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("failed to load details: %w", err)
		}

		for _, d := range rows {
			// UPSERT: qty fisik jadi qty sistem baru; version naik untuk optimistic locking.
			if err := tx.Exec(`
				INSERT INTO stock (product_id, warehouse_id, quantity_system, version, updated_at)
				VALUES (?, ?, ?, 1, now())
				ON CONFLICT (product_id, warehouse_id) DO UPDATE
				SET quantity_system = EXCLUDED.quantity_system,
				    version         = stock.version + 1,
				    updated_at      = now()`,
				d.ProductID, session.WarehouseID, d.StockPhysical).Error; err != nil {
				return fmt.Errorf("failed to write stock for product %s: %w", d.ProductID, err)
			}
		}

		result := tx.Model(&domain.StockOpname{}).
			Where("id = ? AND status = 'completed'", sessionID).
			Updates(map[string]any{
				"status":      "approved",
				"approved_by": approverID,
				"approved_at": time.Now(),
			})
		if result.Error != nil {
			return fmt.Errorf("failed to approve session: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return domain.ErrSessionNotApprovable
		}
		return nil
	})
}

// StockOpnameDetailRepository
type StockOpnameDetailRepository interface {
	Create(ctx context.Context, detail *domain.StockOpnameDetail) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.StockOpnameDetail, error)
	// FindByIdempotencyKey mencari baris hasil retry klien yang sama.
	FindByIdempotencyKey(ctx context.Context, sessionID uuid.UUID, key string) (*domain.StockOpnameDetail, error)
	FindByStockOpnameID(ctx context.Context, stockOpnameID uuid.UUID, limit, offset int) ([]domain.StockOpnameDetail, int64, error)
	CountByStockOpnameID(ctx context.Context, stockOpnameID uuid.UUID) (int64, error)
}

type stockOpnameDetailRepository struct {
	db *gorm.DB
}

func NewStockOpnameDetailRepository(db *gorm.DB) StockOpnameDetailRepository {
	return &stockOpnameDetailRepository{db: db}
}

// stockOpnameDetailInsertFields = kolom yang boleh ditulis saat INSERT.
// `difference` (generated column) dan kolom relasi sengaja tidak ada di sini.
var stockOpnameDetailInsertFields = []string{
	"id", "stock_opname_id", "product_id", "box_id",
	"stock_system", "stock_physical", "input_method", "note",
	"photo_url", "scanned_by", "device_id", "scanned_at", "sync_status",
	"idempotency_key",
}

func (r *stockOpnameDetailRepository) Create(ctx context.Context, detail *domain.StockOpnameDetail) error {
	// Generated column `difference` tidak boleh ditulis oleh aplikasi — DB menghitung
	// secara otomatis. Select field yang diizinkan untuk INSERT.
	if err := r.db.WithContext(ctx).
		Select(stockOpnameDetailInsertFields).
		Create(detail).Error; err != nil {
		// 23505 pada idx unik (stock_opname_id, idempotency_key) = retry klien
		// yang balapan dengan request sebelumnya.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrScanDuplicate
		}
		return fmt.Errorf("failed to create stock opname detail: %w", err)
	}
	return nil
}

func (r *stockOpnameDetailRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockOpnameDetail, error) {
	var detail domain.StockOpnameDetail
	if err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Box").
		Preload("ScannedUser").
		Where("id = ?", id).
		First(&detail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find detail: %w", err)
	}
	return &detail, nil
}

func (r *stockOpnameDetailRepository) FindByIdempotencyKey(ctx context.Context, sessionID uuid.UUID, key string) (*domain.StockOpnameDetail, error) {
	var detail domain.StockOpnameDetail
	if err := r.db.WithContext(ctx).
		Preload("Product").
		Where("stock_opname_id = ? AND idempotency_key = ?", sessionID, key).
		First(&detail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find detail by idempotency key: %w", err)
	}
	return &detail, nil
}

func (r *stockOpnameDetailRepository) FindByStockOpnameID(ctx context.Context, stockOpnameID uuid.UUID, limit, offset int) ([]domain.StockOpnameDetail, int64, error) {
	var details []domain.StockOpnameDetail
	var total int64

	query := r.db.WithContext(ctx).
		Where("stock_opname_id = ?", stockOpnameID)

	if err := query.Model(&domain.StockOpnameDetail{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count details: %w", err)
	}

	if err := query.
		Preload("Product").
		Preload("Box").
		Preload("ScannedUser").
		Limit(limit).
		Offset(offset).
		Order("scanned_at DESC").
		Find(&details).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find details: %w", err)
	}

	return details, total, nil
}

func (r *stockOpnameDetailRepository) CountByStockOpnameID(ctx context.Context, stockOpnameID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.StockOpnameDetail{}).
		Where("stock_opname_id = ?", stockOpnameID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count details: %w", err)
	}
	return count, nil
}

// StockRepository
type StockRepository interface {
	FindByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*domain.Stock, error)
	UpdateWithOptimisticLock(ctx context.Context, stock *domain.Stock) error
	CreateOrUpdate(ctx context.Context, stock *domain.Stock) error
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) FindByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*domain.Stock, error) {
	var stock domain.Stock
	if err := r.db.WithContext(ctx).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find stock: %w", err)
	}
	return &stock, nil
}

func (r *stockRepository) UpdateWithOptimisticLock(ctx context.Context, stock *domain.Stock) error {
	// Optimistic locking: update hanya jika version masih sama
	result := r.db.WithContext(ctx).
		Model(&domain.Stock{}).
		Where("id = ? AND version = ?", stock.ID, stock.Version).
		Updates(map[string]any{
			"quantity_system": stock.QuantitySystem,
			"version":         stock.Version + 1,
			"updated_at":      gorm.Expr("now()"),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update stock: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed: stock was modified by another transaction")
	}

	stock.Version++ // Update version di object
	return nil
}

func (r *stockRepository) CreateOrUpdate(ctx context.Context, stock *domain.Stock) error {
	// Cek apakah stock sudah ada
	existing, err := r.FindByProductAndWarehouse(ctx, stock.ProductID, stock.WarehouseID)
	if err != nil {
		return err
	}

	if existing == nil {
		// Create new
		if err := r.db.WithContext(ctx).Create(stock).Error; err != nil {
			return fmt.Errorf("failed to create stock: %w", err)
		}
		return nil
	}

	// Update existing dengan optimistic lock
	stock.ID = existing.ID
	stock.Version = existing.Version
	return r.UpdateWithOptimisticLock(ctx, stock)
}

// ===== END FILE =====
