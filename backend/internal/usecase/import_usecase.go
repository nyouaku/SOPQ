package usecase

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

// Kolom template import produk. Urutan ini juga urutan yang dibaca ulang saat import,
// jadi template dan parser tak bisa melenceng.
//
// Kolom warehouse_code & stock_system bersifat opsional. Kalau keduanya diisi,
// stok awal produk di gudang tersebut ikut di-upsert ke tabel stock — berguna
// untuk menyiapkan stok awal sebelum SO pertama (SO pertama tak lagi penuh selisih).
var productImportColumns = []string{
	"product_code", "name", "brand", "unit", "minimum_stock", "status",
	"warehouse_code", "stock_system",
}

// maxImportRows membatasi jumlah baris data per file import.
const maxImportRows = 2000

// productStatuses — nilai kolom status yang diterima.
var productStatuses = []string{"active", "inactive", "discontinued"}

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
	productRepo   repository.ProductRepository
	stockRepo     repository.StockRepository
	warehouseRepo repository.WarehouseRepository
}

func NewImportUsecase(
	productRepo repository.ProductRepository,
	stockRepo repository.StockRepository,
	warehouseRepo repository.WarehouseRepository,
) ImportUsecase {
	return &importUsecase{
		productRepo:   productRepo,
		stockRepo:     stockRepo,
		warehouseRepo: warehouseRepo,
	}
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
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, colName, colName, 20)
	}

	// Dropdown status: cegah nilai salah yang bikin update produk tak terpakai.
	lastCol, _ := excelize.ColumnNumberToName(len(productImportColumns))
	statusCol, _ := excelize.ColumnNumberToName(6) // kolom F = status
	dv := excelize.NewDataValidation(true)
	if err := dv.SetDropList(productStatuses); err == nil {
		dv.SetSqref(fmt.Sprintf("%s2:%s%d", statusCol, statusCol, maxImportRows+1))
		_ = f.AddDataValidation(sheet, dv)
	}

	// Baris contoh: kolom wajib diisi, sisanya opsional.
	f.SetCellValue(sheet, "A2", "EL-001")
	f.SetCellValue(sheet, "B2", "Kabel NYA 2.5mm")
	f.SetCellValue(sheet, "C2", "Eterna")
	f.SetCellValue(sheet, "D2", "ROLL")
	f.SetCellValue(sheet, "E2", 10)
	f.SetCellValue(sheet, "F2", "active")
	f.SetCellValue(sheet, "G2", "WH01")
	f.SetCellValue(sheet, "H2", 100)

	// Baris @2 dijadikan contoh saja: hapus sebelum user isi (tapi berguna sbg
	// petunjuk format), jadi ditandai lewat catatan di sheet Petunjuk.

	// Header beku + filter: 2000 baris tetap nyaman digeser.
	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	_ = f.AutoFilter(sheet, fmt.Sprintf("A1:%s1", lastCol), []excelize.AutoFilterOptions{})

	f.NewSheet("Petunjuk")
	hints := [][2]string{
		{"Kolom wajib", "product_code, name"},
		{"Cocok produk berdasarkan", "product_code — kode inilah yang dicetak di QR"},
		{"product_code yang sudah ada", "memperbarui produk tersebut (bukan duplikat)"},
		{"Status", "active (default) / inactive / discontinued — pakai dropdown di kolom F"},
		{"Stok awal (opsional)", "isi warehouse_code + stock_system sekaligus"},
		{"Contoh", "WH01 = kode gudang; stok diisi di baris yang sama"},
		{"Stok yang SUDAH ada", "TIDAK ditimpa — import ulang aman, tak merusak stok riil"},
		{"Batas", fmt.Sprintf("maksimal %d baris data per file, 5 MB", maxImportRows)},
		{"Baris contoh", "baris 2 hanya contoh format — timpa atau hapus sebelum import"},
	}
	for i, h := range hints {
		row := i + 2
		f.SetCellValue("Petunjuk", fmt.Sprintf("A%d", row), h[0])
		f.SetCellValue("Petunjuk", fmt.Sprintf("B%d", row), h[1])
	}
	f.SetCellValue("Petunjuk", "A1", "Bantuan Template Import Produk")
	f.SetCellStyle("Petunjuk", "A1", "B1", headerStyle)
	hintHeader, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle("Petunjuk", fmt.Sprintf("A2:A%d", len(hints)+1), fmt.Sprintf("A2:A%d", len(hints)+1), hintHeader)
	f.SetColWidth("Petunjuk", "A", "A", 32)
	f.SetColWidth("Petunjuk", "B", "B", 80)

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
	// Hitung baris berisi saja — baris kosong di bawah data tak boleh menggagalkan file.
	dataRows := 0
	for _, row := range rows[1:] {
		if !isBlankRow(row) {
			dataRows++
		}
	}
	if dataRows > maxImportRows {
		return nil, fmt.Errorf("%w: maksimal %d baris data per file, file ini %d", domain.ErrImportTooLarge, maxImportRows, dataRows)
	}

	// Peta header → indeks kolom. User sering menambah/geser/buang kolom; membaca
	// berdasarkan posisi membuat data masuk ke kolom yang salah tanpa error.
	colIndex := make(map[string]int, len(rows[0]))
	for i, h := range rows[0] {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, required := range []string{"product_code", "name"} {
		if _, ok := colIndex[required]; !ok {
			return nil, fmt.Errorf("kolom wajib '%s' tidak ditemukan di baris header", required)
		}
	}

	result := &ProductImportResult{Errors: []ImportRowError{}}
	warehouses := u.newWarehouseResolver(ctx)

	for i, row := range rows[1:] {
		rowNum := i + 2
		if isBlankRow(row) {
			continue
		}

		// Baca per nama kolom; kolom yang tak ada di file = string kosong.
		cell := func(name string) string {
			idx, ok := colIndex[name]
			if !ok || idx >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[idx])
		}

		code := cell("product_code")
		name := cell("name")
		unit := cell("unit")

		if code == "" || name == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: "product_code dan name wajib diisi"})
			continue
		}
		if unit == "" {
			unit = "PCS"
		}

		minStock, _ := strconv.Atoi(cell("minimum_stock"))
		status := strings.ToLower(cell("status"))
		if status == "" {
			status = "active"
		}
		if !validProductStatus(status) {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: "status harus salah satu: " + strings.Join(productStatuses, ", ")})
			continue
		}

		// Stok awal (opsional): hanya jika warehouse_code & stock_system keduanya diisi.
		whCode := cell("warehouse_code")
		stockSystemStr := cell("stock_system")
		stockSystem := 0
		hasStock := false
		if whCode != "" && stockSystemStr != "" {
			v, convErr := strconv.Atoi(stockSystemStr)
			if convErr != nil || v < 0 {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: "stock_system harus angka >= 0"})
				continue
			}
			stockSystem = v
			hasStock = true
		}

		existing, err := u.productRepo.FindByCode(ctx, code)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: err.Error()})
			continue
		}

		var productID uuid.UUID
		if existing != nil {
			existing.Name = name
			existing.Brand = cell("brand")
			existing.Unit = unit
			existing.MinimumStock = minStock
			existing.Status = status
			if err := u.productRepo.Update(ctx, existing); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: err.Error()})
				continue
			}
			productID = existing.ID
			result.Updated++
		} else {
			product := &domain.Product{
				ProductCode:  code,
				Name:         name,
				Brand:        cell("brand"),
				Unit:         unit,
				MinimumStock: minStock,
				Status:       status,
			}
			if err := u.productRepo.Create(ctx, product); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: err.Error()})
				continue
			}
			productID = product.ID
			result.Created++
		}

		// Upsert stok awal hanya untuk baris stock yang BELUM ada. Kalau sudah ada,
		// jangan ditimpa — import ulang tak boleh menghancurkan stok riil.
		if hasStock {
			if err := u.upsertInitialStock(ctx, warehouses, productID, whCode, stockSystem, rowNum, result); err != nil {
				continue
			}
		}
	}

	return result, nil
}

func validProductStatus(s string) bool {
	return slices.Contains(productStatuses, s)
}

// warehouseResolver memetakan kode gudang -> ID sekali saja, bukan query per baris.
// Baris yang gudangnya tak dikenal ditolak dengan pesan jelas, bukan error 500.
type warehouseResolver struct {
	byCode map[string]uuid.UUID
	err    error
}

func (u *importUsecase) newWarehouseResolver(ctx context.Context) *warehouseResolver {
	list, _, err := u.warehouseRepo.List(ctx, 1000, 0)
	if err != nil {
		return &warehouseResolver{err: fmt.Errorf("gagal memuat gudang: %w", err)}
	}
	m := make(map[string]uuid.UUID, len(list))
	for _, w := range list {
		m[strings.ToUpper(strings.TrimSpace(w.Code))] = w.ID
	}
	return &warehouseResolver{byCode: m}
}

// upsertInitialStock mengisi stok awal untuk kombinasi produk + gudang. Stok yang
// SUDAH ada tidak ditimpa agar import ulang tidak menghancurkan stok riil.
func (u *importUsecase) upsertInitialStock(
	ctx context.Context,
	wh *warehouseResolver,
	productID uuid.UUID,
	warehouseCode string,
	stockSystem int,
	rowNum int,
	result *ProductImportResult,
) error {
	fail := func(msg string) error {
		result.Failed++
		result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: msg})
		return fmt.Errorf("%s", msg)
	}
	if wh.err != nil {
		return fail(wh.err.Error())
	}
	whID, found := wh.byCode[strings.ToUpper(strings.TrimSpace(warehouseCode))]
	if !found {
		return fail("kode gudang tidak ditemukan: " + warehouseCode)
	}

	// Baris stock sudah ada? Biarkan — jangan menimpa stok riil.
	existingStock, err := u.stockRepo.FindByProductAndWarehouse(ctx, productID, whID)
	if err != nil {
		return fail("gagal cek stok: " + err.Error())
	}
	if existingStock != nil {
		return nil
	}

	newStock := &domain.Stock{
		ProductID:      productID,
		WarehouseID:    whID,
		QuantitySystem: stockSystem,
		Version:        0,
	}
	if err := u.stockRepo.CreateOrUpdate(ctx, newStock); err != nil {
		return fail("gagal simpan stok awal: " + err.Error())
	}
	return nil
}

func isBlankRow(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
