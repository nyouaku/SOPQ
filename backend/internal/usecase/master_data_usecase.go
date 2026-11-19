// ===== FILE: backend/internal/usecase/master_data_usecase.go =====
package usecase

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/google/uuid"
)

// CategoryUsecase
type CategoryUsecase interface {
	CreateCategory(ctx context.Context, category *domain.Category) error
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	ListCategories(ctx context.Context, limit, offset int) ([]domain.Category, int64, error)
}

type categoryUsecase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUsecase(categoryRepo repository.CategoryRepository) CategoryUsecase {
	return &categoryUsecase{categoryRepo: categoryRepo}
}

func (u *categoryUsecase) CreateCategory(ctx context.Context, category *domain.Category) error {
	return u.categoryRepo.Create(ctx, category)
}

func (u *categoryUsecase) GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	category, err := u.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}
	return category, nil
}

func (u *categoryUsecase) UpdateCategory(ctx context.Context, category *domain.Category) error {
	return u.categoryRepo.Update(ctx, category)
}

func (u *categoryUsecase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return u.categoryRepo.Delete(ctx, id)
}

func (u *categoryUsecase) ListCategories(ctx context.Context, limit, offset int) ([]domain.Category, int64, error) {
	return u.categoryRepo.List(ctx, limit, offset)
}

// SupplierUsecase
type SupplierUsecase interface {
	CreateSupplier(ctx context.Context, supplier *domain.Supplier) error
	GetSupplierByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error)
	UpdateSupplier(ctx context.Context, supplier *domain.Supplier) error
	DeleteSupplier(ctx context.Context, id uuid.UUID) error
	ListSuppliers(ctx context.Context, limit, offset int) ([]domain.Supplier, int64, error)
}

type supplierUsecase struct {
	supplierRepo repository.SupplierRepository
}

func NewSupplierUsecase(supplierRepo repository.SupplierRepository) SupplierUsecase {
	return &supplierUsecase{supplierRepo: supplierRepo}
}

func (u *supplierUsecase) CreateSupplier(ctx context.Context, supplier *domain.Supplier) error {
	return u.supplierRepo.Create(ctx, supplier)
}

func (u *supplierUsecase) GetSupplierByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	supplier, err := u.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, fmt.Errorf("supplier not found")
	}
	return supplier, nil
}

func (u *supplierUsecase) UpdateSupplier(ctx context.Context, supplier *domain.Supplier) error {
	return u.supplierRepo.Update(ctx, supplier)
}

func (u *supplierUsecase) DeleteSupplier(ctx context.Context, id uuid.UUID) error {
	return u.supplierRepo.Delete(ctx, id)
}

func (u *supplierUsecase) ListSuppliers(ctx context.Context, limit, offset int) ([]domain.Supplier, int64, error) {
	return u.supplierRepo.List(ctx, limit, offset)
}

// WarehouseUsecase
type WarehouseUsecase interface {
	CreateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error
	GetWarehouseByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	UpdateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error
	DeleteWarehouse(ctx context.Context, id uuid.UUID) error
	ListWarehouses(ctx context.Context, limit, offset int) ([]domain.Warehouse, int64, error)

	CreateLocation(ctx context.Context, location *domain.Location) error
	GetLocationsByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]domain.Location, error)
	UpdateLocation(ctx context.Context, location *domain.Location) error
	DeleteLocation(ctx context.Context, id uuid.UUID) error
}

type warehouseUsecase struct {
	warehouseRepo repository.WarehouseRepository
	locationRepo  repository.LocationRepository
}

func NewWarehouseUsecase(
	warehouseRepo repository.WarehouseRepository,
	locationRepo repository.LocationRepository,
) WarehouseUsecase {
	return &warehouseUsecase{
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
	}
}

func (u *warehouseUsecase) CreateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return u.warehouseRepo.Create(ctx, warehouse)
}

func (u *warehouseUsecase) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	warehouse, err := u.warehouseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warehouse == nil {
		return nil, fmt.Errorf("warehouse not found")
	}
	return warehouse, nil
}

func (u *warehouseUsecase) UpdateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return u.warehouseRepo.Update(ctx, warehouse)
}

func (u *warehouseUsecase) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	return u.warehouseRepo.Delete(ctx, id)
}

func (u *warehouseUsecase) ListWarehouses(ctx context.Context, limit, offset int) ([]domain.Warehouse, int64, error) {
	return u.warehouseRepo.List(ctx, limit, offset)
}

func (u *warehouseUsecase) CreateLocation(ctx context.Context, location *domain.Location) error {
	return u.locationRepo.Create(ctx, location)
}

func (u *warehouseUsecase) GetLocationsByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]domain.Location, error) {
	return u.locationRepo.FindByWarehouseID(ctx, warehouseID)
}

func (u *warehouseUsecase) UpdateLocation(ctx context.Context, location *domain.Location) error {
	return u.locationRepo.Update(ctx, location)
}

func (u *warehouseUsecase) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	return u.locationRepo.Delete(ctx, id)
}
// ===== END FILE =====
