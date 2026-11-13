// ===== FILE: backend/internal/usecase/dashboard_usecase.go =====
package usecase

import (
	"context"
	"fmt"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
)

type DashboardUsecase interface {
	GetSummary(ctx context.Context) (*DashboardSummary, error)
}

type dashboardUsecase struct {
	dashboardRepo repository.DashboardRepository
	activityRepo  repository.ActivityLogRepository
}

func NewDashboardUsecase(
	dashboardRepo repository.DashboardRepository,
	activityRepo repository.ActivityLogRepository,
) DashboardUsecase {
	return &dashboardUsecase{
		dashboardRepo: dashboardRepo,
		activityRepo:  activityRepo,
	}
}

type DashboardSummary struct {
	TotalSKU         int64                    `json:"total_sku"`
	TodayProgress    int64                    `json:"today_progress"`
	TotalDifference  int64                    `json:"total_difference"`
	TotalChecked     int64                    `json:"total_checked"`
	TotalUnchecked   int64                    `json:"total_unchecked"`
	RecentActivities []domain.ActivityLog     `json:"recent_activities"`
}

// recentActivityLimit: jumlah aktivitas terbaru yang ditampilkan di dashboard.
const recentActivityLimit = 10

func (u *dashboardUsecase) GetSummary(ctx context.Context) (*DashboardSummary, error) {
	counts, err := u.dashboardRepo.Counts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load dashboard counts: %w", err)
	}

	recent, _, err := u.activityRepo.List(ctx, repository.ActivityLogFilter{Limit: recentActivityLimit})
	if err != nil {
		return nil, fmt.Errorf("failed to load recent activities: %w", err)
	}

	unchecked := max(counts.TotalSKU-counts.TotalChecked, 0)

	return &DashboardSummary{
		TotalSKU:         counts.TotalSKU,
		TodayProgress:    counts.InProgressScans,
		TotalDifference:  counts.TotalDifference,
		TotalChecked:     counts.TotalChecked,
		TotalUnchecked:   unchecked,
		RecentActivities: recent,
	}, nil
}
