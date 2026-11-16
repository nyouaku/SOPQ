package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

type ExportHandler struct {
	exportUsecase usecase.ExportUsecase
}

func NewExportHandler(exportUsecase usecase.ExportUsecase) *ExportHandler {
	return &ExportHandler{
		exportUsecase: exportUsecase,
	}
}

// ExportStockOpnameExcel godoc
// @Summary Export stock opname report to Excel
// @Description Export detailed stock opname report in Excel format
// @Tags export
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param session_id query string false "Filter by session ID"
// @Success 200 {file} file "Excel file"
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/export/stock-opname [get]
func (h *ExportHandler) ExportStockOpnameExcel(c *gin.Context) {
	filter, err := parseReportFilter(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}

	// session_id kini dibaca parseReportFilter, jadi tidak perlu parsing manual lagi.
	repoFilter := repository.ReportFilter{
		SessionID:   filter.SessionID,
		WarehouseID: filter.WarehouseID,
		SupplierID:  filter.SupplierID,
		UserID:      filter.UserID,
		Brand:       filter.Brand,
		MinDiff:     filter.MinDiff,
		MaxDiff:     filter.MaxDiff,
	}
	if filter.DateFrom != nil {
		s := filter.DateFrom.Format("2006-01-02")
		repoFilter.DateFrom = &s
	}
	if filter.DateTo != nil {
		s := filter.DateTo.Format("2006-01-02")
		repoFilter.DateTo = &s
	}

	file, filename, err := h.exportUsecase.ExportStockOpnameToExcel(c.Request.Context(), repoFilter)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "EXPORT_FAILED", err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")

	if err := file.Write(c.Writer); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "WRITE_FAILED", err.Error())
		return
	}
}

// ExportProductsExcel godoc
// @Summary Export products to Excel
// @Description Export all products in Excel format
// @Tags export
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file "Excel file"
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/export/products [get]
func (h *ExportHandler) ExportProductsExcel(c *gin.Context) {
	file, filename, err := h.exportUsecase.ExportProductsToExcel(c.Request.Context())
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "EXPORT_FAILED", err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")

	if err := file.Write(c.Writer); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "WRITE_FAILED", err.Error())
		return
	}
}
