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
	// ProgressByBox mengembalikan progress scan per rak/box dalam satu sesi:
	// jumlah produk di-scan vs target produk di box tsb.
	ProgressByBox(ctx context.Context, sessionID uuid.UUID) ([]domain.BoxProgress, error)
	// FindProductTrend mengembalikan riwayat stok fisik satu produk di seluruh
	// sesi (dalam rentang waktu opsional). Dipakai untuk grafik tren stok.
	FindProductTrend(ctx context.Context, productID uuid.UUID, from, to *time.Time, limit int) ([]domain.ProductStockTrend, error)
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

// ProgressByBox mengembalikan progress scan per rak/box: jumlah produk sudah
// di-scan vs target (jumlah produk yg box_id-nya sama). Target dihitung dari
// tabel products (bukan hanya yg sudah ter-scan), agar bar yang belum ter-scan
// tetap muncul dgn 0%.
func (r *stockOpnameDetailRepository) ProgressByBox(ctx context.Context, sessionID uuid.UUID) ([]domain.BoxProgress, error) {
	var rows []domain.BoxProgress
	err := r.db.WithContext(ctx).Raw(`
		WITH scanned AS (
			SELECT box_id, COUNT(*) AS scanned
			FROM stock_opname_detail
			WHERE stock_opname_id = ? AND box_id IS NOT NULL
			GROUP BY box_id
		),
		targets AS (
			SELECT b.id AS box_id, b.code AS box_code, b.name AS box_name,
				COUNT(p.id) AS target
			FROM boxes b
			JOIN products p ON p.box_id = b.id AND p.deleted_at IS NULL
			JOIN locations l ON l.id = b.location_id
			JOIN stock_opname so ON so.warehouse_id = l.warehouse_id
			WHERE so.id = ?
			GROUP BY b.id, b.code, b.name
		)
		SELECT t.box_id, t.box_code, t.box_name, t.target,
			COALESCE(s.scanned, 0)::int AS scanned,
			ROUND(COALESCE(s.scanned,0)::numeric * 100.0 / NULLIF(t.target,0), 1)::float8 AS percent
		FROM targets t
		LEFT JOIN scanned s ON s.box_id = t.box_id
		ORDER BY t.box_code ASC
	`, sessionID, sessionID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load progress by box: %w", err)
	}
	return rows, nil
}

// StockRepository
type StockRepository interface {
	FindByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*domain.Stock, error)
	UpdateWithOptimisticLock(ctx context.Context, stock *domain.Stock) error
	CreateOrUpdate(ctx context.Context, stock *domain.Stock) error
	// ReconcileWarehouses menghitung rekap stok setiap produk di seluruh gudang
	// yang dipilih. Hanya mengembalikan produk yang setidaknya ada di 1 gudang.
	ReconcileWarehouses(ctx context.Context, warehouseIDs []uuid.UUID) ([]domain.WarehouseReconciliation, error)
	// AdjustStock mengubah quantity_system dan mengembalikan nilai sebelum & sesudah.
	// Menggunakan optimistic locking untuk keamanan.
	AdjustStock(ctx context.Context, productID, warehouseID uuid.UUID, delta int, version int) (oldVal, newVal int, err error)
	// CountProductsInWarehouse menghitung jumlah produk yang punya baris stock di
	// gudang tsb — dipakai sebagai TARGET stock opname sesi di gudang itu. Produk
	// tanpa baris stock di gudang tsb bukan target (tak dihitung).
	CountProductsInWarehouse(ctx context.Context, warehouseID uuid.UUID) (int, error)
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

// CountProductsInWarehouse menghitung target stock opname: jumlah produk yang
// punya baris stock (apa pun quantity_system-nya) di gudang tsb.
func (r *stockRepository) CountProductsInWarehouse(ctx context.Context, warehouseID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Stock{}).
		Where("warehouse_id = ?", warehouseID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count products in warehouse: %w", err)
	}
	return int(count), nil
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

func (r *stockOpnameDetailRepository) FindProductTrend(ctx context.Context, productID uuid.UUID, from, to *time.Time, limit int) ([]domain.ProductStockTrend, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.db.WithContext(ctx).
		Table("stock_opname_detail AS d").
		Select(`d.scanned_at, d.stock_system, d.stock_physical, d.difference, so.code AS session_code`).
		Joins("JOIN stock_opname so ON so.id = d.stock_opname_id").
		Where("d.product_id = ?", productID).
		Order("d.scanned_at ASC").
		Limit(limit)
	if from != nil {
		q = q.Where("d.scanned_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("d.scanned_at <= ?", *to)
	}
	var rows []domain.ProductStockTrend
	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch product trend: %w", err)
	}
	return rows, nil
}

// AdjustStock mengubah quantity_system dengan delta (bisa negatif) menggunakan
// optimistic locking. Mengembalikan nilai sebelum & sesudah untuk audit log.
func (r *stockRepository) AdjustStock(ctx context.Context, productID, warehouseID uuid.UUID, delta int, version int) (oldVal, newVal int, err error) {
	stock, err := r.FindByProductAndWarehouse(ctx, productID, warehouseID)
	if err != nil {
		return 0, 0, err
	}
	if stock == nil {
		return 0, 0, fmt.Errorf("stock not found for product in this warehouse")
	}

	oldVal = stock.QuantitySystem
	newVal = oldVal + delta
	if newVal < 0 {
		return 0, 0, fmt.Errorf("adjustment would make stock negative (current %d, delta %d)", oldVal, delta)
	}

	stock.QuantitySystem = newVal
	stock.Version = version
	if err := r.UpdateWithOptimisticLock(ctx, stock); err != nil {
		return 0, 0, err
	}
	return oldVal, newVal, nil
}

// ReconcileWarehouses menghitung rekap stok per produk di seluruh gudang yang
// dipilih. Produk dianggap "hilang" dari gudang jika tidak ada baris di tabel
// stock untuk kombinasi produk + gudang tersebut.
func (r *stockRepository) ReconcileWarehouses(ctx context.Context, warehouseIDs []uuid.UUID) ([]domain.WarehouseReconciliation, error) {
	if len(warehouseIDs) == 0 {
		return nil, nil
	}

	// Nama gudang untuk pelengkap missing_from
	var whs []struct {
		ID   uuid.UUID `gorm:"column:id"`
		Name string    `gorm:"column:name"`
	}
	if err := r.db.WithContext(ctx).
		Table("warehouses").
		Select("id, name").
		Where("id IN ?", warehouseIDs).
		Find(&whs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch warehouses: %w", err)
	}

	// Ambil stok semua produk di gudang-gudang terpilih.
	var rows []struct {
		ProductID      uuid.UUID `gorm:"column:product_id"`
		ProductCode    string    `gorm:"column:product_code"`
		ProductName    string    `gorm:"column:product_name"`
		WarehouseID    uuid.UUID `gorm:"column:warehouse_id"`
		WarehouseName  string    `gorm:"column:warehouse_name"`
		QuantitySystem int       `gorm:"column:quantity_system"`
	}
	if err := r.db.WithContext(ctx).
		Table("stock AS s").
		Select(`s.product_id, p.product_code, p.name AS product_name,
			s.warehouse_id, w.name AS warehouse_name, s.quantity_system`).
		Joins("JOIN products p ON p.id = s.product_id").
		Joins("JOIN warehouses w ON w.id = s.warehouse_id").
		Where("s.warehouse_id IN ?", warehouseIDs).
		Order("p.product_code ASC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch reconciliation rows: %w", err)
	}

	whNameByID := make(map[uuid.UUID]string, len(whs))
	for _, w := range whs {
		whNameByID[w.ID] = w.Name
	}
	whIDList := make([]uuid.UUID, 0, len(whs))
	whNameList := make([]string, 0, len(whs))
	for _, w := range whs {
		whIDList = append(whIDList, w.ID)
		whNameList = append(whNameList, w.Name)
	}

	// Group by product.
	type acc struct {
		code     string
		name     string
		total    int
		present  int
		perWH    map[uuid.UUID]bool
	}
	accMap := make(map[uuid.UUID]*acc)
	for _, row := range rows {
		a, ok := accMap[row.ProductID]
		if !ok {
			a = &acc{
				code:  row.ProductCode,
				name:  row.ProductName,
				perWH: make(map[uuid.UUID]bool),
			}
			accMap[row.ProductID] = a
		}
		a.total += row.QuantitySystem
		if _, exists := a.perWH[row.WarehouseID]; !exists {
			a.perWH[row.WarehouseID] = true
			a.present++
		}
	}

	results := make([]domain.WarehouseReconciliation, 0, len(accMap))
	for productID, a := range accMap {
		missing := make([]string, 0)
		for i, wid := range whIDList {
			if _, ok := a.perWH[wid]; !ok {
				missing = append(missing, whNameByID[wid])
			}
			_ = i
			_ = whNameList
		}
		results = append(results, domain.WarehouseReconciliation{
			ProductID:       productID,
			ProductCode:     a.code,
			ProductName:     a.name,
			QuantitySystem:  a.total,
			TotalAll:        a.total,
			PresentInCount:  a.present,
			TotalWarehouses: len(whIDList),
			MissingFrom:     missing,
		})
	}

	return results, nil
}
