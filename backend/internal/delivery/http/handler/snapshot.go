// ===== FILE: backend/internal/delivery/http/handler/snapshot.go =====
package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"gorm.io/gorm"
)

type SnapshotHandler struct {
	db *gorm.DB
}

func NewSnapshotHandler(db *gorm.DB) *SnapshotHandler {
	return &SnapshotHandler{db: db}
}

// GetSnapshotComparison godoc
// @Summary Compare current stock vs a past snapshot
// @Description Returns per-product stock delta between now and a snapshot date
// @Tags snapshots
// @Produce json
// @Security BearerAuth
// @Param days query int false "Days ago for baseline (default 30)" default(30)
// @Param warehouse_id query string false "Filter by warehouse"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/snapshots/compare [get]
func (h *SnapshotHandler) GetSnapshotComparison(c *gin.Context) {
	days := 30
	if raw := c.Query("days"); raw != "" {
		fmt.Sscanf(raw, "%d", &days)
	}
	if days < 1 || days > 365 {
		days = 30
	}

	snapshotTime := time.Now().AddDate(0, 0, -days)

	warehouseID := c.Query("warehouse_id")

	// Current stock (from products join stock)
	type currentRow struct {
		ProductID      uuid.UUID `gorm:"column:product_id"`
		ProductCode    string    `gorm:"column:product_code"`
		ProductName    string    `gorm:"column:product_name"`
		CurrentStock   int       `gorm:"column:current_stock"`
		MinStock       int       `gorm:"column:min_stock"`
		WarehouseName  string    `gorm:"column:warehouse_name"`
	}

	// Historical stock approximation from stock_opname_detail (latest scan before snapshotTime)
	type historyRow struct {
		ProductID      uuid.UUID `gorm:"column:product_id"`
		HistoricalStock int      `gorm:"column:historical_stock"`
	}

	// Current stock query
	currentQuery := h.db.WithContext(c.Request.Context()).
		Table("stock AS s").
		Select(`s.product_id, p.product_code, p.name AS product_name,
			s.quantity_system AS current_stock, p.minimum_stock AS min_stock,
			w.name AS warehouse_name`).
		Joins("JOIN products p ON p.id = s.product_id").
		Joins("JOIN warehouses w ON w.id = s.warehouse_id")

	if warehouseID != "" {
		currentQuery = currentQuery.Where("s.warehouse_id = ?", warehouseID)
	}

	var current []currentRow
	if err := currentQuery.Find(&current).Error; err != nil {
		utils.RespondInternalError(c, "Failed to load current stock: "+err.Error())
		return
	}

	// Historical stock: ambil stok_fisik terakhir sebelum snapshotTime per produk
	historyQuery := h.db.WithContext(c.Request.Context()).
		Table("stock_opname_detail AS d").
		Select(`DISTINCT ON (d.product_id) d.product_id, d.stock_physical AS historical_stock`).
		Where("d.scanned_at <= ?", snapshotTime).
		Order("d.product_id, d.scanned_at DESC")

	var history []historyRow
	if err := historyQuery.Find(&history).Error; err != nil {
		utils.RespondInternalError(c, "Failed to load historical stock: "+err.Error())
		return
	}

	// Build history map
	historyMap := make(map[uuid.UUID]int)
	for _, h := range history {
		historyMap[h.ProductID] = h.HistoricalStock
	}

	// Build comparison
	type comparisonItem struct {
		ProductID       uuid.UUID `json:"product_id"`
		ProductCode     string    `json:"product_code"`
		ProductName     string    `json:"product_name"`
		WarehouseName   string    `json:"warehouse_name"`
		SnapshotStock   int       `json:"snapshot_stock"`
		CurrentStock    int       `json:"current_stock"`
		Change          int       `json:"change"`
		ChangePercent   float64   `json:"change_percent"`
		MinStock        int       `json:"min_stock"`
		Status          string    `json:"status"` // ok, low, critical, increased
	}

	items := make([]comparisonItem, 0, len(current))
	for _, c := range current {
		histStock := historyMap[c.ProductID]
		change := c.CurrentStock - histStock
		var changePercent float64
		if histStock > 0 {
			changePercent = float64(change) / float64(histStock) * 100.0
		} else if change > 0 {
			changePercent = 100.0
		}

		status := "ok"
		if change < 0 && c.CurrentStock < c.MinStock {
			status = "critical"
		} else if c.CurrentStock < c.MinStock {
			status = "low"
		} else if change > 0 {
			status = "increased"
		}

		items = append(items, comparisonItem{
			ProductID:     c.ProductID,
			ProductCode:   c.ProductCode,
			ProductName:   c.ProductName,
			WarehouseName: c.WarehouseName,
			SnapshotStock: histStock,
			CurrentStock:  c.CurrentStock,
			Change:        change,
			ChangePercent: changePercent,
			MinStock:      c.MinStock,
			Status:        status,
		})
	}

	// Sort by change (descending - biggest changes first)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if abs(items[j].Change) > abs(items[i].Change) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Summary stats
	var totalIncreased, totalDecreased, totalUnchanged int
	for _, item := range items {
		if item.Change > 0 {
			totalIncreased++
		} else if item.Change < 0 {
			totalDecreased++
		} else {
			totalUnchanged++
		}
	}

	utils.RespondOK(c, gin.H{
		"snapshot_days":    days,
		"snapshot_date":    snapshotTime.Format(time.RFC3339),
		"total_products":   len(items),
		"total_increased":  totalIncreased,
		"total_decreased":  totalDecreased,
		"total_unchanged":  totalUnchanged,
		"comparisons":      items,
	})
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
