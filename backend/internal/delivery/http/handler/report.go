// ===== FILE: backend/internal/delivery/http/handler/report.go =====
package handler

import (
	"net/http"
	"strconv"

	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/logger"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportHandler struct {
	reportUsecase       usecase.ReportUsecase
	notificationUsecase usecase.NotificationUsecase
}

func NewReportHandler(
	reportUsecase usecase.ReportUsecase,
	notificationUsecase usecase.NotificationUsecase,
) *ReportHandler {
	return &ReportHandler{
		reportUsecase:       reportUsecase,
		notificationUsecase: notificationUsecase,
	}
}

// GetStockOpnameReport godoc
// @Summary Get stock opname report
// @Description Get stock opname report with filters (supervisor/admin)
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string false "Filter by warehouse"
// @Param brand query string false "Filter by brand"
// @Param supplier_id query string false "Filter by supplier"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param user_id query string false "Filter by user"
// @Param min_diff query int false "Minimum difference"
// @Param max_diff query int false "Maximum difference"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/reports/stock-opname [get]
func (h *ReportHandler) GetStockOpnameReport(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	filter, err := parseReportFilter(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}
	filter.Limit = limit
	filter.Offset = offset

	res, err := h.reportUsecase.GetStockOpnameReport(c.Request.Context(), filter)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"data":    res.Rows,
			"summary": res.Summary,
		"total":   res.Total,
		"page":    page,
		"limit":   limit,
	})
}

// ExportStockOpnameReport godoc
// @Summary Export stock opname report
// @Description Export report to CSV (supervisor/admin)
// @Tags reports
// @Produce text/csv
// @Security BearerAuth
// @Param warehouse_id query string false "Filter by warehouse"
// @Param brand query string false "Filter by brand"
// @Param supplier_id query string false "Filter by supplier"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {file} file
// @Router /api/v1/reports/stock-opname/export [get]
func (h *ReportHandler) ExportStockOpnameReport(c *gin.Context) {
	filter, err := parseReportFilter(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}

	// Header dikirim sebelum body ditulis. Error setelah ini tidak bisa
	// diubah jadi 500, jadi usecase hanya menulis baris yang sudah lolos.
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=stock_opname_report.csv")

	if err := h.reportUsecase.ExportStockOpnameCSV(c.Request.Context(), filter, c.Writer); err != nil {
		logger.Error("CSV export gagal di tengah stream: %v", err)
	}
}

// GetActivityLogs godoc
// @Summary Get activity logs
// @Description Get activity log history (supervisor/admin)
// @Tags activity-logs
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "Filter by user"
// @Param action query string false "Filter by action"
// @Param entity_type query string false "Filter by entity type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/activity-logs [get]
func (h *ReportHandler) GetActivityLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	filter := repository.ActivityLogFilter{
		Action:     c.Query("action"),
		EntityType: c.Query("entity_type"),
		Limit:      limit,
		Offset:     offset,
	}

	if userID := c.Query("user_id"); userID != "" {
		id, err := uuid.Parse(userID)
		if err == nil {
			filter.UserID = &id
		}
	}

	logs, total, err := h.reportUsecase.GetActivityLogs(c.Request.Context(), filter)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetNotifications godoc
// @Summary Get notifications
// @Description Get user notifications
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/notifications [get]
func (h *ReportHandler) GetNotifications(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.RespondUnauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	notifications, total, err := h.notificationUsecase.ListNotifications(c.Request.Context(), &userID, limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// MarkNotificationAsRead godoc
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/notifications/{id}/read [patch]
func (h *ReportHandler) MarkNotificationAsRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid notification ID")
		return
	}

	if err := h.notificationUsecase.MarkAsRead(c.Request.Context(), id); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "MARK_READ_FAILED", err.Error())
		return
	}

	utils.RespondOK(c, gin.H{"message": "Notification marked as read"})
}
// ===== END FILE =====
