// ===== FILE: backend/internal/repository/activity_log_repository.go =====
package repository

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityLogRepository interface {
	Create(ctx context.Context, log *domain.ActivityLog) error
	List(ctx context.Context, filter ActivityLogFilter) ([]domain.ActivityLog, int64, error)
}

type ActivityLogFilter struct {
	UserID     *uuid.UUID
	Action     string
	EntityType string
	Limit      int
	Offset     int
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) Create(ctx context.Context, log *domain.ActivityLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create activity log: %w", err)
	}
	return nil
}

func (r *activityLogRepository) List(ctx context.Context, filter ActivityLogFilter) ([]domain.ActivityLog, int64, error) {
	var logs []domain.ActivityLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.ActivityLog{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count activity logs: %w", err)
	}

	if err := query.
		Preload("User").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list activity logs: %w", err)
	}

	return logs, total, nil
}
// ===== END FILE =====
