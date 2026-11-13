// ===== FILE: backend/internal/repository/notification_repository.go =====
package repository

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]domain.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	var notification domain.Notification
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}
	return &notification, nil
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Notification{})

	if userID != nil {
		query = query.Where("user_id = ? OR user_id IS NULL", userID)
	} else {
		query = query.Where("user_id IS NULL") // broadcast only
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	if err := query.
		Preload("User").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	return notifications, total, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("failed to mark all as read: %w", err)
	}
	return nil
}
// ===== END FILE =====
