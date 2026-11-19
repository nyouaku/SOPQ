package usecase

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

// importProductRepo menambah penyimpanan nyata di atas mock yang sudah ada,
// supaya import bisa diuji end-to-end tanpa Postgres.
type importProductRepo struct {
	*mockProductRepository
}

func newImportProductRepo() *importProductRepo {
	return &importProductRepo{&mockProductRepository{products: map[string]*domain.Product{}}}
}

func (f *importProductRepo) Create(ctx context.Context, p *domain.Product) error {
	p.ID = uuid.New()
	f.products[p.ProductCode] = p
	return nil
}

func (f *importProductRepo) Update(ctx context.Context, p *domain.Product) error {
	f.products[p.ProductCode] = p
	return nil
}

func (f *importProductRepo) FindByCode(ctx context.Context, code string) (*domain.Product, error) {
	return f.products[code], nil
}

type fakeWarehouseRepo struct{ list []domain.Warehouse }

func (f *fakeWarehouseRepo) List(ctx context.Context, limit, offset int) ([]domain.Warehouse, int64, error) {
	return f.list, int64(len(f.list)), nil
}
func (f *fakeWarehouseRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	return nil, nil
}
func (f *fakeWarehouseRepo) Create(ctx context.Context, w *domain.Warehouse) error { return nil }
func (f *fakeWarehouseRepo) Update(ctx context.Context, w *domain.Warehouse) error { return nil }
func (f *fakeWarehouseRepo) Delete(ctx context.Context, id uuid.UUID) error        { return nil }

var _ repository.WarehouseRepository = (*fakeWarehouseRepo)(nil)

// buildXLSX membuat file Excel dari grid string (baris x kolom).
func buildXLSX(t *testing.T, rows [][]string) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, row := range rows {
		for j, v := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}
	return buf
}

func newImportUsecase() (*importUsecase, *importProductRepo, *mockStockRepository) {
	p := newImportProductRepo()
	s := &mockStockRepository{stocks: map[uuid.UUID]*domain.Stock{}}
	wh := &fakeWarehouseRepo{list: []domain.Warehouse{{ID: uuid.New(), Code: "WH01"}}}
	return NewImportUsecase(p, s, wh).(*importUsecase), p, s
}

// Regresi: kolom dibaca berdasarkan NAMA header, bukan posisi. File dengan kolom
// diurut ulang/buang tetap ter-import benar, tidak ada data masuk kolom salah.
func TestImportProducts_ReadsColumnsByHeaderName(t *testing.T) {
	uc, products, _ := newImportUsecase()
	file := buildXLSX(t, [][]string{
		{"name", "unit", "product_code", "brand", "status"},
		{"Kabel NYA", "ROLL", "EL-001", "Eterna", "inactive"},
	})

	res, err := uc.ImportProducts(context.Background(), file)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if res.Created != 1 || res.Failed != 0 {
		t.Fatalf("created=%d failed=%d errs=%v", res.Created, res.Failed, res.Errors)
	}
	p := products.products["EL-001"]
	if p == nil {
		t.Fatal("EL-001 tidak tersimpan")
	}
	if p.Name != "Kabel NYA" || p.Unit != "ROLL" || p.Brand != "Eterna" {
		t.Fatalf("kolom tertukar: %+v", p)
	}
	if p.Status != "inactive" {
		t.Fatalf("status salah: %q", p.Status)
	}
}

// Baris kosong di bawah data tidak boleh menambah hitungan atau menggagalkan import.
func TestImportProducts_IgnoresTrailingBlankRows(t *testing.T) {
	uc, _, _ := newImportUsecase()
	file := buildXLSX(t, [][]string{
		{"product_code", "name"},
		{"EL-001", "Kabel"},
		{"", ""},
		{"", ""},
	})

	res, err := uc.ImportProducts(context.Background(), file)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if res.Created != 1 || res.Failed != 0 {
		t.Fatalf("created=%d failed=%d", res.Created, res.Failed)
	}
}

// Status di luar daftar ditolak dengan pesan jelas (bukan diam-diam jadi 'active').
func TestImportProducts_RejectsUnknownStatus(t *testing.T) {
	uc, _, _ := newImportUsecase()
	file := buildXLSX(t, [][]string{
		{"product_code", "name", "status"},
		{"EL-001", "Kabel", "rusak"},
	})

	res, err := uc.ImportProducts(context.Background(), file)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if res.Failed != 1 || res.Created != 0 {
		t.Fatalf("failed=%d created=%d", res.Failed, res.Created)
	}
	if !strings.Contains(res.Errors[0].Message, "status") {
		t.Fatalf("pesan error tak menyebut status: %v", res.Errors[0].Message)
	}
}

// Header wajib hilang -> tolak file, jangan import sebagian dengan kolom salah.
func TestImportProducts_MissingRequiredHeader(t *testing.T) {
	uc, _, _ := newImportUsecase()
	file := buildXLSX(t, [][]string{
		{"nama_produk", "brand"},
		{"Kabel", "Eterna"},
	})

	if _, err := uc.ImportProducts(context.Background(), file); err == nil {
		t.Fatal("harusnya error: kolom product_code tidak ada")
	}
}

// Stok awal yang SUDAH ada tidak ditimpa — import ulang tak boleh merusak stok riil.
func TestImportProducts_DoesNotOverwriteExistingStock(t *testing.T) {
	uc, products, stocks := newImportUsecase()
	stockOf := func() *domain.Stock {
		for _, st := range stocks.stocks {
			return st
		}
		return nil
	}

	first := buildXLSX(t, [][]string{
		{"product_code", "name", "warehouse_code", "stock_system"},
		{"EL-001", "Kabel", "WH01", "100"},
	})
	if _, err := uc.ImportProducts(context.Background(), first); err != nil {
		t.Fatalf("import pertama: %v", err)
	}
	if got := stockOf().QuantitySystem; got != 100 {
		t.Fatalf("stok awal salah: %d", got)
	}

	second := buildXLSX(t, [][]string{
		{"product_code", "name", "warehouse_code", "stock_system"},
		{"EL-001", "Kabel Baru", "WH01", "5"},
	})
	if _, err := uc.ImportProducts(context.Background(), second); err != nil {
		t.Fatalf("import kedua: %v", err)
	}
	if got := stockOf().QuantitySystem; got != 100 {
		t.Fatalf("stok riil tertimpa: %d, harusnya tetap 100", got)
	}
	if products.products["EL-001"].Name != "Kabel Baru" {
		t.Fatal("update produk ikut gagal")
	}
}