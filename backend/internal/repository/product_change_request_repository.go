// ===== FILE: backend/internal/repository/product_change_request_repository.go =====
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"gorm.io/gorm"
)

type ChangeRequestFilter struct {
	Status string
	Limit  int
	Offset int
}

type ProductChangeRequestRepository interface {
	Create(ctx context.Context, req *domain.ProductChangeRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.ProductChangeRequest, error)
	List(ctx context.Context, f ChangeRequestFilter) ([]domain.ProductChangeRequest, int64, error)
	Update(ctx context.Context, req *domain.ProductChangeRequest) error
}

type productChangeRequestRepository struct {
	db *gorm.DB
}

func NewProductChangeRequestRepository(db *gorm.DB) ProductChangeRequestRepository {
	return &productChangeRequestRepository{db: db}
}

func (r *productChangeRequestRepository) Create(ctx context.Context, req *domain.ProductChangeRequest) error {
	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return fmt.Errorf("failed to create product change request: %w", err)
	}
	return nil
}

func (r *productChangeRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ProductChangeRequest, error) {
	var req domain.ProductChangeRequest
	if err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Requester").
		Preload("Reviewer").
		Where("id = ?", id).
		First(&req).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find product change request: %w", err)
	}
	return &req, nil
}

func (r *productChangeRequestRepository) List(ctx context.Context, f ChangeRequestFilter) ([]domain.ProductChangeRequest, int64, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	q := r.db.WithContext(ctx).Model(&domain.ProductChangeRequest{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count change requests: %w", err)
	}
	var rows []domain.ProductChangeRequest
	if err := q.
		Preload("Product").
		Preload("Requester").
		Preload("Reviewer").
		Order("created_at DESC").
		Limit(f.Limit).Offset(f.Offset).
		Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list change requests: %w", err)
	}
	return rows, total, nil
}

func (r *productChangeRequestRepository) Update(ctx context.Context, req *domain.ProductChangeRequest) error {
	if err := r.db.WithContext(ctx).Save(req).Error; err != nil {
		return fmt.Errorf("failed to update product change request: %w", err)
	}
	return nil
}
