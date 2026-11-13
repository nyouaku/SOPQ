// ===== FILE: backend/internal/repository/dashboard_repository.go =====
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// DashboardCounts adalah agregat dashboard, diambil dengan tiga query SQL
// (bukan loop per sesi dan per produk).
type DashboardCounts struct {
	TotalSKU        int64
	TotalChecked    int64
	TotalDifference int64
	// InProgressScans: jumlah detail pada sesi berstatus in_progress.
	// Perilaku sama dengan versi lama (bukan "hari ini"), agar angka tidak berubah diam-diam.
	InProgressScans int64
}

type DashboardRepository interface {
	Counts(ctx context.Context) (DashboardCounts, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) Counts(ctx context.Context) (DashboardCounts, error) {
	var c DashboardCounts
	db := r.db.WithContext(ctx)

	// Query 1: total produk aktif.
	if err := db.Raw(`SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`).
		Scan(&c.TotalSKU).Error; err != nil {
		return c, fmt.Errorf("failed to count products: %w", err)
	}

	// Query 2: detail dari sesi completed/in_progress — total dicek, selisih, dan scan in_progress.
	if err := db.Raw(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE d.difference <> 0),
			COUNT(*) FILTER (WHERE so.status = 'in_progress')
		FROM stock_opname_detail d
		JOIN stock_opname so ON so.id = d.stock_opname_id
		WHERE so.status IN ('in_progress', 'completed')`).
		Row().Scan(&c.TotalChecked, &c.TotalDifference, &c.InProgressScans); err != nil {
		return c, fmt.Errorf("failed to aggregate details: %w", err)
	}

	return c, nil
}

// ===== END FILE =====
