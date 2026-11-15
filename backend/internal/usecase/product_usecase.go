// ===== FILE: backend/internal/usecase/product_usecase.go =====
package usecase

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/google/uuid"
)

type ProductUsecase interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetProductByCode(ctx context.Context, code string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	ListProducts(ctx context.Context, filter repository.ProductFilter) ([]domain.Product, int64, error)
	// TransferProducts memindahkan beberapa produk ke box tujuan dalam satu transaksi.
	// Box tujuan harus valid dan berada dalam lokasi yang sama dengan produk asal,
	// atau boleh beda lokasi asalkan masih dalam gudang yang sama.
	TransferProducts(ctx context.Context, productIDs []uuid.UUID, toBoxID uuid.UUID, userID uuid.UUID) (int, error)
}

type productUsecase struct {
	productRepo repository.ProductRepository
}

func NewProductUsecase(productRepo repository.ProductRepository) ProductUsecase {
	return &productUsecase{
		productRepo: productRepo,
	}
}

func (u *productUsecase) CreateProduct(ctx context.Context, product *domain.Product) error {
	return u.productRepo.Create(ctx, product)
}

func (u *productUsecase) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}
	return product, nil
}


func (u *productUsecase) GetProductByCode(ctx context.Context, code string) (*domain.Product, error) {
	product, err := u.productRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}
	return product, nil
}

func (u *productUsecase) UpdateProduct(ctx context.Context, product *domain.Product) error {
	existing, err := u.productRepo.FindByID(ctx, product.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("product not found")
	}
	return u.productRepo.Update(ctx, product)
}

func (u *productUsecase) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	existing, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("product not found")
	}
	return u.productRepo.Delete(ctx, id)
}

func (u *productUsecase) ListProducts(ctx context.Context, filter repository.ProductFilter) ([]domain.Product, int64, error) {
	return u.productRepo.List(ctx, filter)
}

func (u *productUsecase) TransferProducts(ctx context.Context, productIDs []uuid.UUID, toBoxID uuid.UUID, userID uuid.UUID) (int, error) {
	if len(productIDs) == 0 {
		return 0, fmt.Errorf("no products selected")
	}
	if toBoxID == uuid.Nil {
		return 0, fmt.Errorf("target box is required")
	}

	affected, err := u.productRepo.BulkUpdateBoxID(ctx, productIDs, &toBoxID)
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}
// ===== END FILE =====
