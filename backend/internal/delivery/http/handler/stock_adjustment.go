// ===== FILE: backend/internal/delivery/http/handler/stock_adjustment.go =====
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/nyouaku/SOPQ/backend/pkg/validator"
)

type StockAdjustmentHandler struct {
	stockOpnameUsecase usecase.StockOpnameUsecase
}

func NewStockAdjustmentHandler(stockOpnameUsecase usecase.StockOpnameUsecase) *StockAdjustmentHandler {
	return &StockAdjustmentHandler{stockOpnameUsecase: stockOpnameUsecase}
}

// AdjustStock godoc
// @Summary Manual stock adjustment
// @Description Adjust stock system manually with mandatory reason
// @Tags stocks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body AdjustStockRequest true "Adjustment data"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stocks/adjust [post]
func (h *StockAdjustmentHandler) AdjustStock(c *gin.Context) {
	var req AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}
	delta, err := strconv.Atoi(req.Delta)
	if err != nil {
		utils.RespondBadRequest(c, "Delta must be an integer")
		return
	}

	stock, err := h.stockOpnameUsecase.AdjustStock(c.Request.Context(), usecase.AdjustStockInput{
		ProductID:   productID,
		WarehouseID: warehouseID,
		Delta:       delta,
		Reason:      req.Reason,
		Note:        req.Note,
		UserID:      userID,
		IPAddress:   c.ClientIP(),
	})
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, stock)
}

// AdjustStockRequest — DTO untuk penyesuaian stok.
type AdjustStockRequest struct {
	ProductID   string `json:"product_id" validate:"required,uuid"`
	WarehouseID string `json:"warehouse_id" validate:"required,uuid"`
	Delta       string `json:"delta" validate:"required"`
	Reason      string `json:"reason" validate:"required,oneof=damaged lost expired found typo_fix other"`
	Note        string `json:"note"`
}

// ===== END FILE =====
