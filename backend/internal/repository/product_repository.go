// ===== FILE: backend/internal/repository/product_repository.go =====
package repository

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	FindByCode(ctx context.Context, code string) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ProductFilter) ([]domain.Product, int64, error)
}

type ProductFilter struct {
	Search     string
	SupplierID *uuid.UUID
	Status     string
	Limit      int
	Offset     int
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *productRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Box").
		Preload("Box.Location").
		Preload("Box.Location.Warehouse").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}
	return &product, nil
}


func (r *productRepository) FindByCode(ctx context.Context, code string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Box").
		Preload("Box.Location").
		Preload("Box.Location.Warehouse").
		Where("product_code = ? AND deleted_at IS NULL", code).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find product by code: %w", err)
	}
	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete
	now := gorm.DeletedAt{Valid: true}
	if err := r.db.WithContext(ctx).
		Model(&domain.Product{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error; err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

func (r *productRepository) List(ctx context.Context, filter ProductFilter) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Product{}).
		Where("deleted_at IS NULL")

	// Apply filters
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR product_code ILIKE ?", searchPattern, searchPattern)
	}
	if filter.SupplierID != nil {
		query = query.Where("supplier_id = ?", filter.SupplierID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get data with pagination
	if err := query.
		Preload("Supplier").
		Preload("Box").
		Preload("Box.Location").
		Preload("Box.Location.Warehouse").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order("created_at DESC").
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}
// ===== END FILE =====
