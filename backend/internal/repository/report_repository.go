// ===== FILE: backend/internal/repository/report_repository.go =====
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportFilter adalah filter laporan opname. Dipakai bersama oleh JSON dan CSV/Excel
// supaya ketiganya selalu menyaring dengan aturan yang sama.
type ReportFilter struct {
	SessionID   *uuid.UUID // stock_opname.id; dipakai ekspor Excel per sesi
	WarehouseID *uuid.UUID
	SupplierID  *uuid.UUID
	UserID      *uuid.UUID
	Brand       string
	DateFrom    *string // YYYY-MM-DD, inklusif
	DateTo      *string // YYYY-MM-DD, inklusif
	MinDiff     *int
	MaxDiff     *int
	Limit       int
	Offset      int
}

// ReportRow adalah satu baris hasil join. Field sengaja datar supaya mudah diekspor.
type ReportRow struct {
	SessionCode   string
	WarehouseName string
	ProductCode   string
	ProductName   string
	Brand         string
	Category      string
	Supplier      string
	StockSystem   int
	StockPhysical int
	Difference    int
	InputMethod   string
	ScannedBy     string
	ScannedAt     string
}

// ReportSummary adalah agregat seluruh baris yang cocok filter, bukan hanya halaman aktif.
type ReportSummary struct {
	TotalCount    int64
	TotalSystem   int64
	TotalPhysical int64
	TotalDiff     int64
	MismatchCount int64
}

type ReportRepository interface {
	// Query mengembalikan halaman baris, total baris yang cocok, dan agregat seluruh hasil.
	Query(ctx context.Context, f ReportFilter) ([]ReportRow, int64, ReportSummary, error)
	// Stream memanggil fn untuk setiap baris yang cocok, tanpa pagination. Dipakai
	// ekspor CSV/Excel agar tidak memuat seluruh hasil ke memori.
	Stream(ctx context.Context, f ReportFilter, fn func(ReportRow) error) error
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

// baseQuery memuat join dan semua filter. Dipakai bersama oleh Query dan Stream.
func (r *reportRepository) baseQuery(ctx context.Context, f ReportFilter) *gorm.DB {
	q := r.db.WithContext(ctx).
		Table("stock_opname_detail d").
		Joins("JOIN stock_opname so ON so.id = d.stock_opname_id").
		Joins("JOIN warehouses w ON w.id = so.warehouse_id").
		Joins("JOIN products p ON p.id = d.product_id").
		Joins("LEFT JOIN suppliers s ON s.id = p.supplier_id").
		Joins("LEFT JOIN users u ON u.id = d.scanned_by")

	if f.SessionID != nil {
		q = q.Where("d.stock_opname_id = ?", *f.SessionID)
	}
	if f.WarehouseID != nil {
		q = q.Where("so.warehouse_id = ?", *f.WarehouseID)
	}
	if f.SupplierID != nil {
		q = q.Where("p.supplier_id = ?", *f.SupplierID)
	}
	if f.UserID != nil {
		q = q.Where("d.scanned_by = ?", *f.UserID)
	}
	if f.Brand != "" {
		q = q.Where("p.brand ILIKE ?", "%"+f.Brand+"%")
	}
	if f.DateFrom != nil {
		q = q.Where("d.scanned_at >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		// Batas atas inklusif untuk seluruh hari tersebut.
		q = q.Where("d.scanned_at < (?::date + INTERVAL '1 day')", *f.DateTo)
	}
	// Selisih = stock_physical - stock_system (kolom generated `difference`).
	if f.MinDiff != nil {
		q = q.Where("d.difference >= ?", *f.MinDiff)
	}
	if f.MaxDiff != nil {
		q = q.Where("d.difference <= ?", *f.MaxDiff)
	}
	return q
}

const selectColumns = `
	so.code AS session_code,
	w.name AS warehouse_name,
	p.product_code AS product_code,
	p.name AS product_name,
	COALESCE(p.brand, '') AS brand,
	COALESCE(s.name, '') AS supplier,
	d.stock_system AS stock_system,
	d.stock_physical AS stock_physical,
	d.difference AS difference,
	d.input_method AS input_method,
	COALESCE(u.full_name, '') AS scanned_by,
	TO_CHAR(d.scanned_at, 'YYYY-MM-DD HH24:MI:SS') AS scanned_at`

func (r *reportRepository) Query(ctx context.Context, f ReportFilter) ([]ReportRow, int64, ReportSummary, error) {
	// Satu query untuk total + agregat: hitung seluruh hasil filter, bukan halaman.
	var sum ReportSummary
	if err := r.baseQuery(ctx, f).
		Select(`COUNT(*) AS total,
			COALESCE(SUM(d.stock_system), 0) AS total_system,
			COALESCE(SUM(d.stock_physical), 0) AS total_physical,
			COALESCE(SUM(d.difference), 0) AS total_diff,
			COUNT(*) FILTER (WHERE d.difference <> 0) AS mismatch_count`).
		Scan(&sum).Error; err != nil {
		return nil, 0, ReportSummary{}, fmt.Errorf("failed to aggregate report rows: %w", err)
	}

	var rows []ReportRow
	if err := r.baseQuery(ctx, f).
		Select(selectColumns).
		Order("d.scanned_at DESC, d.id DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, ReportSummary{}, fmt.Errorf("failed to query report rows: %w", err)
	}
	return rows, sum.TotalCount, sum, nil
}

func (r *reportRepository) Stream(ctx context.Context, f ReportFilter, fn func(ReportRow) error) error {
	rows, err := r.baseQuery(ctx, f).
		Select(selectColumns).
		Order("d.scanned_at DESC, d.id DESC").
		Rows()
	if err != nil {
		return fmt.Errorf("failed to stream report rows: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row ReportRow
		if err := r.db.ScanRows(rows, &row); err != nil {
			return fmt.Errorf("failed to scan report row: %w", err)
		}
		if err := fn(row); err != nil {
			return err
		}
	}
	return rows.Err()
}

// ===== END FILE =====
