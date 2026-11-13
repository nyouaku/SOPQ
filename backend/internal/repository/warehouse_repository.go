// ===== FILE: backend/internal/repository/warehouse_repository.go =====
package repository

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *domain.Warehouse) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	Update(ctx context.Context, warehouse *domain.Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]domain.Warehouse, int64, error)
}

type warehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) WarehouseRepository {
	return &warehouseRepository{db: db}
}

func (r *warehouseRepository) Create(ctx context.Context, warehouse *domain.Warehouse) error {
	if err := r.db.WithContext(ctx).Create(warehouse).Error; err != nil {
		return fmt.Errorf("failed to create warehouse: %w", err)
	}
	return nil
}

func (r *warehouseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&warehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find warehouse: %w", err)
	}
	return &warehouse, nil
}

func (r *warehouseRepository) Update(ctx context.Context, warehouse *domain.Warehouse) error {
	if err := r.db.WithContext(ctx).Save(warehouse).Error; err != nil {
		return fmt.Errorf("failed to update warehouse: %w", err)
	}
	return nil
}

func (r *warehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Warehouse{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete warehouse: %w", err)
	}
	return nil
}

func (r *warehouseRepository) List(ctx context.Context, limit, offset int) ([]domain.Warehouse, int64, error) {
	var warehouses []domain.Warehouse
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Warehouse{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count warehouses: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("name ASC").
		Find(&warehouses).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list warehouses: %w", err)
	}

	return warehouses, total, nil
}

// LocationRepository
type LocationRepository interface {
	Create(ctx context.Context, location *domain.Location) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]domain.Location, error)
	Update(ctx context.Context, location *domain.Location) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type locationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) Create(ctx context.Context, location *domain.Location) error {
	if err := r.db.WithContext(ctx).Create(location).Error; err != nil {
		return fmt.Errorf("failed to create location: %w", err)
	}
	return nil
}

func (r *locationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	var location domain.Location
	if err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Where("id = ?", id).
		First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find location: %w", err)
	}
	return &location, nil
}

func (r *locationRepository) FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]domain.Location, error) {
	var locations []domain.Location
	if err := r.db.WithContext(ctx).
		Where("warehouse_id = ?", warehouseID).
		Order("code ASC").
		Find(&locations).Error; err != nil {
		return nil, fmt.Errorf("failed to find locations by warehouse: %w", err)
	}
	return locations, nil
}

func (r *locationRepository) Update(ctx context.Context, location *domain.Location) error {
	if err := r.db.WithContext(ctx).Save(location).Error; err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}
	return nil
}

func (r *locationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Location{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete location: %w", err)
	}
	return nil
}

// BoxRepository
type BoxRepository interface {
	Create(ctx context.Context, box *domain.Box) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Box, error)
	FindByLocationID(ctx context.Context, locationID uuid.UUID) ([]domain.Box, error)
	Update(ctx context.Context, box *domain.Box) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type boxRepository struct {
	db *gorm.DB
}

func NewBoxRepository(db *gorm.DB) BoxRepository {
	return &boxRepository{db: db}
}

func (r *boxRepository) Create(ctx context.Context, box *domain.Box) error {
	if err := r.db.WithContext(ctx).Create(box).Error; err != nil {
		return fmt.Errorf("failed to create box: %w", err)
	}
	return nil
}

func (r *boxRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Box, error) {
	var box domain.Box
	if err := r.db.WithContext(ctx).
		Preload("Location").
		Preload("Location.Warehouse").
		Where("id = ?", id).
		First(&box).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find box: %w", err)
	}
	return &box, nil
}

func (r *boxRepository) FindByLocationID(ctx context.Context, locationID uuid.UUID) ([]domain.Box, error) {
	var boxes []domain.Box
	if err := r.db.WithContext(ctx).
		Where("location_id = ?", locationID).
		Order("code ASC").
		Find(&boxes).Error; err != nil {
		return nil, fmt.Errorf("failed to find boxes by location: %w", err)
	}
	return boxes, nil
}

func (r *boxRepository) Update(ctx context.Context, box *domain.Box) error {
	if err := r.db.WithContext(ctx).Save(box).Error; err != nil {
		return fmt.Errorf("failed to update box: %w", err)
	}
	return nil
}

func (r *boxRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Box{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete box: %w", err)
	}
	return nil
}
// ===== END FILE =====
