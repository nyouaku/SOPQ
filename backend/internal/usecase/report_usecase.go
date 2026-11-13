// ===== FILE: backend/internal/usecase/report_usecase.go =====
package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
)

type ReportUsecase interface {
	GetStockOpnameReport(ctx context.Context, filter StockOpnameReportFilter) (StockOpnameReportResult, error)
	ExportStockOpnameCSV(ctx context.Context, filter StockOpnameReportFilter, writer io.Writer) error
	GetActivityLogs(ctx context.Context, filter repository.ActivityLogFilter) ([]domain.ActivityLog, int64, error)
}

type reportUsecase struct {
	stockOpnameDetailRepo repository.StockOpnameDetailRepository
	activityLogRepo       repository.ActivityLogRepository
	reportRepo            repository.ReportRepository
}

func NewReportUsecase(
	stockOpnameDetailRepo repository.StockOpnameDetailRepository,
	activityLogRepo repository.ActivityLogRepository,
	reportRepo repository.ReportRepository,
) ReportUsecase {
	return &reportUsecase{
		stockOpnameDetailRepo: stockOpnameDetailRepo,
		activityLogRepo:       activityLogRepo,
		reportRepo:            reportRepo,
	}
}

type StockOpnameReportFilter struct {
	WarehouseID  *uuid.UUID
	Brand        string
	SupplierID   *uuid.UUID
	DateFrom     *time.Time
	DateTo       *time.Time
	UserID       *uuid.UUID
	MinDiff      *int
	MaxDiff      *int
	Limit        int
	Offset       int
}

// StockOpnameReportResult: halaman baris + total + agregat seluruh hasil filter.
type StockOpnameReportResult struct {
	Rows    []StockOpnameReportRow
	Total   int64
	Summary StockOpnameReportSummary
}

type StockOpnameReportSummary struct {
	TotalSystem   int64 `json:"total_system"`
	TotalPhysical int64 `json:"total_physical"`
	TotalDiff     int64 `json:"total_difference"`
	MismatchCount int64 `json:"mismatch_count"`
}

type StockOpnameReportRow struct {
	SessionCode   string    `json:"session_code"`
	WarehouseName string    `json:"warehouse_name"`
	ProductCode   string    `json:"product_code"`
	ProductName   string    `json:"product_name"`
	Brand         string    `json:"brand"`
	Supplier      string    `json:"supplier"`
	StockSystem   int       `json:"stock_system"`
	StockPhysical int       `json:"stock_physical"`
	Difference    int       `json:"difference"`
	InputMethod   string    `json:"input_method"`
	ScannedBy     string    `json:"scanned_by"`
	ScannedAt     time.Time `json:"scanned_at"`
}

// toRepoFilter mengubah filter usecase menjadi filter repository.
func toRepoFilter(f StockOpnameReportFilter) repository.ReportFilter {
	rf := repository.ReportFilter{
		WarehouseID: f.WarehouseID,
		SupplierID:  f.SupplierID,
		UserID:      f.UserID,
		Brand:       f.Brand,
		MinDiff:     f.MinDiff,
		MaxDiff:     f.MaxDiff,
		Limit:       f.Limit,
		Offset:      f.Offset,
	}
	if f.DateFrom != nil {
		s := f.DateFrom.Format("2006-01-02")
		rf.DateFrom = &s
	}
	if f.DateTo != nil {
		s := f.DateTo.Format("2006-01-02")
		rf.DateTo = &s
	}
	return rf
}

func toReportRow(r repository.ReportRow) StockOpnameReportRow {
	scannedAt, _ := time.Parse("2006-01-02 15:04:05", r.ScannedAt)
	return StockOpnameReportRow{
		SessionCode:   r.SessionCode,
		WarehouseName: r.WarehouseName,
		ProductCode:   r.ProductCode,
		ProductName:   r.ProductName,
		Brand:         r.Brand,
		Supplier:      r.Supplier,
		StockSystem:   r.StockSystem,
		StockPhysical: r.StockPhysical,
		Difference:    r.Difference,
		InputMethod:   r.InputMethod,
		ScannedBy:     r.ScannedBy,
		ScannedAt:     scannedAt,
	}
}

func (u *reportUsecase) GetStockOpnameReport(ctx context.Context, filter StockOpnameReportFilter) (StockOpnameReportResult, error) {
	rows, total, sum, err := u.reportRepo.Query(ctx, toRepoFilter(filter))
	if err != nil {
		return StockOpnameReportResult{}, fmt.Errorf("failed to load stock opname report: %w", err)
	}
	out := make([]StockOpnameReportRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, toReportRow(r))
	}
	return StockOpnameReportResult{
		Rows:  out,
		Total: total,
		Summary: StockOpnameReportSummary{
			TotalSystem:   sum.TotalSystem,
			TotalPhysical: sum.TotalPhysical,
			TotalDiff:     sum.TotalDiff,
			MismatchCount: sum.MismatchCount,
		},
	}, nil
}

// maxExportRows membatasi ekspor CSV. Batas ini wajib: tanpanya satu request bisa
// menahan koneksi DB dan menulis jutaan baris ke response.
const maxExportRows = 100000

// ExportStockOpnameCSV men-stream baris laporan ke writer satu per satu. Ia tidak
// memuat seluruh hasil ke memori, dan melaporkan error bila jumlah baris melebihi batas.
func (u *reportUsecase) ExportStockOpnameCSV(ctx context.Context, filter StockOpnameReportFilter, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{
		"Session Code", "Warehouse", "Product Code", "Product Name",
		"Brand", "Supplier", "Stock System", "Stock Physical",
		"Difference", "Input Method", "Scanned By", "Scanned At",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	count := 0
	err := u.reportRepo.Stream(ctx, toRepoFilter(filter), func(r repository.ReportRow) error {
		count++
		if count > maxExportRows {
			return fmt.Errorf("%w: maksimal %d baris per ekspor, persempit filter", domain.ErrExportTooLarge, maxExportRows)
		}
		return csvWriter.Write([]string{
			r.SessionCode,
			r.WarehouseName,
			r.ProductCode,
			r.ProductName,
			r.Brand,
			r.Supplier,
			strconv.Itoa(r.StockSystem),
			strconv.Itoa(r.StockPhysical),
			strconv.Itoa(r.Difference),
			r.InputMethod,
			r.ScannedBy,
			r.ScannedAt,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to stream report: %w", err)
	}
	return nil
}

func (u *reportUsecase) GetActivityLogs(ctx context.Context, filter repository.ActivityLogFilter) ([]domain.ActivityLog, int64, error) {
	return u.activityLogRepo.List(ctx, filter)
}
// ===== END FILE =====
