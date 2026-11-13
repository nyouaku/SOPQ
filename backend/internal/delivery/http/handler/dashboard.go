// ===== FILE: backend/internal/delivery/http/handler/dashboard.go =====
package handler

import (
	"net/http"

	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardUsecase usecase.DashboardUsecase
}

func NewDashboardHandler(dashboardUsecase usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{
		dashboardUsecase: dashboardUsecase,
	}
}

// GetSummary godoc
// @Summary Get dashboard summary
// @Description Get summary statistics for dashboard
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=usecase.DashboardSummary}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/dashboard/summary [get]
func (h *DashboardHandler) GetSummary(c *gin.Context) {
	summary, err := h.dashboardUsecase.GetSummary(c.Request.Context())
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "FAILED", err.Error())
		return
	}

	utils.RespondOK(c, summary)
}
