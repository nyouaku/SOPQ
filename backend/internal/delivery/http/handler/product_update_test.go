package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
)

// fakeProductUsecase mengisi usecase.ProductUsecase; hanya GetProductByID & UpdateProduct dipakai.
type fakeProductUsecase struct {
	stored  domain.Product
	updated *domain.Product
}

func (f *fakeProductUsecase) CreateProduct(context.Context, *domain.Product) error { return nil }
func (f *fakeProductUsecase) GetProductByID(_ context.Context, _ uuid.UUID) (*domain.Product, error) {
	cp := f.stored
	return &cp, nil
}
func (f *fakeProductUsecase) GetProductByBarcode(context.Context, string) (*domain.Product, error) {
	return nil, nil
}
func (f *fakeProductUsecase) GetProductByCode(context.Context, string) (*domain.Product, error) {
	return nil, nil
}
func (f *fakeProductUsecase) UpdateProduct(_ context.Context, p *domain.Product) error {
	f.updated = p
	return nil
}
func (f *fakeProductUsecase) DeleteProduct(context.Context, uuid.UUID) error { return nil }
func (f *fakeProductUsecase) ListProducts(context.Context, repository.ProductFilter) ([]domain.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeProductUsecase) TransferProducts(_ context.Context, _ []uuid.UUID, _ uuid.UUID, _ uuid.UUID) (int, error) {
	return 0, nil
}

// Regresi: edit produk tanpa mengirim barcode/product_code tidak boleh mengosongkannya
// (sebelumnya Save() menulis barcode = '' lalu bentrok products_barcode_key).
func TestUpdateProductKeepsIdentityFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	fake := &fakeProductUsecase{stored: domain.Product{
		ID: id, ProductCode: "PRD-1", Name: "Lama", Unit: "pcs", Status: "active",
	}}
	h := NewProductHandler(fake)

	body, _ := json.Marshal(map[string]any{"name": "Baru", "unit": "pcs", "minimum_stock": 5})
	r := gin.New()
	r.PUT("/products/:id", h.UpdateProduct)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/products/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if fake.updated == nil {
		t.Fatal("UpdateProduct not called")
	}
	if fake.updated.ProductCode != "PRD-1" {
		t.Fatalf("identity overwritten: code=%q", fake.updated.ProductCode)
	}
	if fake.updated.Name != "Baru" || fake.updated.MinimumStock != 5 {
		t.Fatalf("edit not applied: name=%q min=%d", fake.updated.Name, fake.updated.MinimumStock)
	}
}
