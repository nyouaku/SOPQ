// ===== FILE: backend/internal/delivery/http/handler/reconciliation.go =====
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

type ReconciliationHandler struct {
	stockOpnameUsecase usecase.StockOpnameUsecase
}

func NewReconciliationHandler(stockOpnameUsecase usecase.StockOpnameUsecase) *ReconciliationHandler {
	return &ReconciliationHandler{stockOpnameUsecase: stockOpnameUsecase}
}

// ReconcileWarehouses godoc
// @Summary Multi-warehouse stock reconciliation
// @Description Compares product stock across selected warehouses
// @Tags reconciliation
// @Produce json
// @Security BearerAuth
// @Param warehouse_ids body []string true "Warehouse IDs"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/reconciliation [post]
func (h *ReconciliationHandler) ReconcileWarehouses(c *gin.Context) {
	var req struct {
		WarehouseIDs []string `json:"warehouse_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}
	if len(req.WarehouseIDs) < 2 {
		utils.RespondBadRequest(c, "At least 2 warehouse IDs required")
		return
	}

	ids := make([]uuid.UUID, 0, len(req.WarehouseIDs))
	for _, raw := range req.WarehouseIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			utils.RespondBadRequest(c, "Invalid warehouse ID: "+raw)
			return
		}
		ids = append(ids, id)
	}

	result, err := h.stockOpnameUsecase.ReconcileWarehouses(c.Request.Context(), ids)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"reconciliation": result,
		"total_products": len(result),
	})
}

// ===== END FILE =====
