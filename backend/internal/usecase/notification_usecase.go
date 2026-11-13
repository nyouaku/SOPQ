// ===== FILE: backend/internal/usecase/notification_usecase.go =====
package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationUsecase interface {
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	ListNotifications(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]domain.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CheckAndNotifyDifference(ctx context.Context, stockSystem, stockPhysical int, productName string, userID uuid.UUID, thresholdPercent int) error
}

type notificationUsecase struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationUsecase(notificationRepo repository.NotificationRepository) NotificationUsecase {
	return &notificationUsecase{
		notificationRepo: notificationRepo,
	}
}

func (u *notificationUsecase) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	return u.notificationRepo.Create(ctx, notification)
}

func (u *notificationUsecase) ListNotifications(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
	return u.notificationRepo.ListByUser(ctx, userID, limit, offset)
}

func (u *notificationUsecase) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return u.notificationRepo.MarkAsRead(ctx, id)
}

func (u *notificationUsecase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return u.notificationRepo.MarkAllAsRead(ctx, userID)
}

// CheckAndNotifyDifference memeriksa selisih dan membuat notifikasi jika > threshold
func (u *notificationUsecase) CheckAndNotifyDifference(ctx context.Context, stockSystem, stockPhysical int, productName string, userID uuid.UUID, thresholdPercent int) error {
	if stockSystem == 0 {
		return nil // Skip jika stock sistem 0
	}

	difference := stockPhysical - stockSystem
	percentDiff := (float64(difference) / float64(stockSystem)) * 100

	// Cek apakah selisih melebihi threshold (bisa positif atau negatif)
	if percentDiff > float64(thresholdPercent) || percentDiff < -float64(thresholdPercent) {
		// json.Marshal, bukan Sprintf: nama produk bisa memuat kutip / backslash
		// dan Sprintf akan menghasilkan jsonb yang tidak sah.
		metadata, err := json.Marshal(map[string]any{
			"product_name":   productName,
			"stock_system":   stockSystem,
			"stock_physical": stockPhysical,
			"difference":     difference,
			"percent_diff":   percentDiff,
		})
		if err != nil {
			return fmt.Errorf("failed to build notification metadata: %w", err)
		}

		notification := &domain.Notification{
			UserID:   &userID,
			Type:     "diff_over_threshold",
			Title:    "Selisih Stock Melebihi Threshold",
			Message:  fmt.Sprintf("Produk '%s' memiliki selisih %.1f%% (sistem: %d, fisik: %d)", productName, percentDiff, stockSystem, stockPhysical),
			Metadata: datatypes.JSON(metadata),
			IsRead:   false,
		}

		if err := u.notificationRepo.Create(ctx, notification); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	return nil
}
// ===== END FILE =====
