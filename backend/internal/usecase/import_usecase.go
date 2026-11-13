package usecase

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

// Kolom template import produk. Urutan ini juga urutan yang dibaca ulang saat import,
// jadi template dan parser tak bisa melenceng.
var productImportColumns = []string{
	"product_code", "name", "brand", "unit", "minimum_stock", "status",
}

// maxImportRows membatasi jumlah baris data per file import.
const maxImportRows = 2000

type ImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type ProductImportResult struct {
	Created int              `json:"created"`
	Updated int              `json:"updated"`
	Failed  int              `json:"failed"`
	Errors  []ImportRowError `json:"errors"`
}

type ImportUsecase interface {
	ProductImportTemplate() (*excelize.File, string, error)
	ImportProducts(ctx context.Context, r io.Reader) (*ProductImportResult, error)
}

type importUsecase struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

func NewImportUsecase(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
) ImportUsecase {
	return &importUsecase{productRepo: productRepo, categoryRepo: categoryRepo}
}

func (u *importUsecase) ProductImportTemplate() (*excelize.File, string, error) {
	f := excelize.NewFile()
	sheet := "Products"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"2E7D32"}, Pattern: 1},
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create style: %w", err)
	}

	for i, col := range productImportColumns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		f.SetColWidth(sheet, string(rune('A'+i)), string(rune('A'+i)), 20)
	}

	// Baris contoh: kolom wajib diisi, sisanya opsional.
	f.SetCellValue(sheet, "A2", "EL-001")
	f.SetCellValue(sheet, "B2", "Kabel NYA 2.5mm")
	f.SetCellValue(sheet, "C2", "8991234567001")
	f.SetCellValue(sheet, "D2", "Eterna")
	f.SetCellValue(sheet, "E2", "kabel")
	f.SetCellValue(sheet, "F2", "ROLL")
	f.SetCellValue(sheet, "G2", "Elektronik & Listrik")
	f.SetCellValue(sheet, "H2", 10)
	f.SetCellValue(sheet, "I2", "active")

	f.NewSheet("Petunjuk")
	f.SetCellValue("Petunjuk", "A1", "Kolom wajib: product_code, name, unit")
	f.SetCellValue("Petunjuk", "A2", "Produk dicocokkan berdasarkan product_code (kode yang dicetak di QR)")
	f.SetCellValue("Petunjuk", "A3", "product_code yang sudah ada akan memperbarui produk tersebut")
	f.SetColWidth("Petunjuk", "A", "A", 80)

	return f, "template-import-products.xlsx", nil
}

func (u *importUsecase) ImportProducts(ctx context.Context, r io.Reader) (*ProductImportResult, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("file tidak bisa dibaca sebagai Excel: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("file tidak punya sheet")
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca sheet: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("file kosong, tidak ada baris data")
	}
	// Batas baris: impor besar memakan memori dan mengunci tabel produk lama.
	if len(rows)-1 > maxImportRows {
		return nil, fmt.Errorf("%w: maksimal %d baris data per file, file ini %d", domain.ErrImportTooLarge, maxImportRows, len(rows)-1)
	}

	// Peta nama kategori → ID, sekali query saja.
	categories, _, err := u.categoryRepo.List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat kategori: %w", err)
	}
	categoryByName := make(map[string]uuid.UUID, len(categories))
	for _, c := range categories {
		categoryByName[strings.ToLower(strings.TrimSpace(c.Name))] = c.ID
	}

	result := &ProductImportResult{Errors: []ImportRowError{}}

	for i, row := range rows[1:] {
		rowNum := i + 2
		if isBlankRow(row) {
			continue
		}

		cell := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		code := cell(0)
		name := cell(1)
		unit := cell(3)

		if code == "" || name == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: "product_code dan name wajib diisi"})
			continue
		}
		if unit == "" {
			unit = "PCS"
		}

		minStock, _ := strconv.Atoi(cell(4))
		status := cell(5)
		if status == "" {
			status = "active"
		}

		existing, err := u.productRepo.FindByCode(ctx, code)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: err.Error()})
			continue
		}

		if existing != nil {
			existing.Name = name
			existing.Brand = cell(2)
			existing.Unit = unit
			existing.MinimumStock = minStock
			existing.Status = status
			if err := u.productRepo.Update(ctx, existing); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: err.Error()})
				continue
			}
			result.Updated++
			continue
		}

		product := &domain.Product{
			ProductCode:  code,
			Name:         name,
			Brand:        cell(2),
			Unit:         unit,
			MinimumStock: minStock,
			Status:       status,
		}
		if err := u.productRepo.Create(ctx, product); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: err.Error()})
			continue
		}
		result.Created++
	}

	return result, nil
}

func isBlankRow(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
