// ===== FILE: backend/internal/usecase/export_usecase.go =====
package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

// maxExcelExportRows membatasi ekspor Excel. Excel memuat seluruh baris di memori
// per workbook, jadi batas ini melindungi server dari ekspor yang tak terkendali.
const maxExcelExportRows = 50000

type ExportUsecase interface {
	ExportStockOpnameToExcel(ctx context.Context, filter repository.ReportFilter) (*excelize.File, string, error)
	ExportProductsToExcel(ctx context.Context) (*excelize.File, string, error)
}

type exportUsecase struct {
	stockOpnameDetailRepo repository.StockOpnameDetailRepository
	productRepo           repository.ProductRepository
	reportRepo            repository.ReportRepository
}

func NewExportUsecase(
	stockOpnameDetailRepo repository.StockOpnameDetailRepository,
	productRepo repository.ProductRepository,
	reportRepo repository.ReportRepository,
) ExportUsecase {
	return &exportUsecase{
		stockOpnameDetailRepo: stockOpnameDetailRepo,
		productRepo:           productRepo,
		reportRepo:            reportRepo,
	}
}

// ExportStockOpnameToExcel menulis baris laporan yang cocok filter ke workbook.
// Memakai excelize.StreamWriter: baris ditulis berurutan tanpa menahan seluruh sheet
// di memori. Wajib SetRow berurutan dari baris 1 ke bawah.
func (u *exportUsecase) ExportStockOpnameToExcel(ctx context.Context, filter repository.ReportFilter) (*excelize.File, string, error) {
	f := excelize.NewFile()
	sheet := "Stock Opname Report"
	if _, err := f.NewSheet(sheet); err != nil {
		f.Close()
		return nil, "", fmt.Errorf("failed to create sheet: %w", err)
	}
	f.DeleteSheet("Sheet1")

	sw, err := f.NewStreamWriter(sheet)
	if err != nil {
		f.Close()
		return nil, "", fmt.Errorf("failed to create stream writer: %w", err)
	}

	headers := []any{"No", "Scanned At", "Session Code", "Product Code", "Product Name",
		"Brand", "Supplier", "Warehouse", "Stock System", "Stock Physical",
		"Difference", "Input Method", "Scanned By"}
	if err := sw.SetRow("A1", headers); err != nil {
		f.Close()
		return nil, "", fmt.Errorf("failed to write header: %w", err)
	}

	row := 2
	err = u.reportRepo.Stream(ctx, filter, func(r repository.ReportRow) error {
		if row-1 > maxExcelExportRows {
			return fmt.Errorf("%w: maksimal %d baris per ekspor, persempit filter", domain.ErrExportTooLarge, maxExcelExportRows)
		}
		cell := "A" + strconv.Itoa(row)
		values := []any{
			row - 1, r.ScannedAt, r.SessionCode, r.ProductCode, r.ProductName,
			r.Brand, r.Supplier, r.WarehouseName, r.StockSystem, r.StockPhysical,
			r.Difference, r.InputMethod, r.ScannedBy,
		}
		if err := sw.SetRow(cell, values); err != nil {
			return fmt.Errorf("failed to write row %d: %w", row, err)
		}
		row++
		return nil
	})
	if err != nil {
		f.Close()
		return nil, "", fmt.Errorf("failed to stream report: %w", err)
	}

	if err := sw.Flush(); err != nil {
		f.Close()
		return nil, "", fmt.Errorf("failed to flush stream writer: %w", err)
	}

	filename := fmt.Sprintf("stock-opname-report-%s.xlsx", time.Now().Format("20060102-150405"))
	return f, filename, nil
}

func (u *exportUsecase) ExportProductsToExcel(ctx context.Context) (*excelize.File, string, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	sheetName := "Products"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"2E7D32"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Set headers
	headers := []string{"No", "SKU", "Name", "Brand", "Supplier", "Unit", "Status"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 5)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 35)
	f.SetColWidth(sheetName, "D", "D", 18)
	f.SetColWidth(sheetName, "E", "E", 15)
	f.SetColWidth(sheetName, "F", "F", 15)
	f.SetColWidth(sheetName, "G", "G", 20)
	f.SetColWidth(sheetName, "H", "H", 10)
	f.SetColWidth(sheetName, "I", "I", 10)

	// Fetch products
	products, _, err := u.productRepo.List(ctx, repository.ProductFilter{
		Limit:  9999,
		Offset: 0,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch products: %w", err)
	}

	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "D3D3D3", Style: 1},
			{Type: "top", Color: "D3D3D3", Style: 1},
			{Type: "bottom", Color: "D3D3D3", Style: 1},
			{Type: "right", Color: "D3D3D3", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})

	// Write data
	row := 2
	for i, product := range products {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), product.ProductCode)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), product.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), product.Brand)

		supplierName := ""
		if product.Supplier != nil {
			supplierName = product.Supplier.Name
		}
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), supplierName)

		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), product.Unit)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), product.Status)

		// Apply style
		for col := 1; col <= 9; col++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}

		row++
	}

	// Freeze first row
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	filename := fmt.Sprintf("products-%s.xlsx", time.Now().Format("20060102-150405"))
	return f, filename, nil
}
