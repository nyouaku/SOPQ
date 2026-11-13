// ===== FILE: backend/internal/repository/supplier_repository.go =====
package repository

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *domain.Supplier) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error)
	Update(ctx context.Context, supplier *domain.Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]domain.Supplier, int64, error)
}

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	if err := r.db.WithContext(ctx).Create(supplier).Error; err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}
	return nil
}

func (r *supplierRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&supplier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find supplier: %w", err)
	}
	return &supplier, nil
}

func (r *supplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	if err := r.db.WithContext(ctx).Save(supplier).Error; err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}
	return nil
}

func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Supplier{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}
	return nil
}

func (r *supplierRepository) List(ctx context.Context, limit, offset int) ([]domain.Supplier, int64, error) {
	var suppliers []domain.Supplier
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Supplier{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count suppliers: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("name ASC").
		Find(&suppliers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list suppliers: %w", err)
	}

	return suppliers, total, nil
}
// ===== END FILE =====
