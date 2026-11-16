// ===== FILE: backend/internal/delivery/http/handler/stock_ledger.go =====
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

type StockLedgerHandler struct {
	ledgerUsecase usecase.StockLedgerUsecase
}

func NewStockLedgerHandler(ledgerUsecase usecase.StockLedgerUsecase) *StockLedgerHandler {
	return &StockLedgerHandler{ledgerUsecase: ledgerUsecase}
}

// GetProductLedger godoc
// @Summary Get product stock ledger
// @Description Returns full stock movement history for a product in a warehouse
// @Tags stock-ledger
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID"
// @Param warehouse_id path string true "Warehouse ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stock-ledger/product/{product_id}/{warehouse_id} [get]
func (h *StockLedgerHandler) GetProductLedger(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}
	warehouseID, err := uuid.Parse(c.Param("warehouse_id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	entries, total, err := h.ledgerUsecase.GetProductLedger(c.Request.Context(), productID, warehouseID, limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"ledger": entries,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// GetWarehouseLedger godoc
// @Summary Get warehouse stock ledger
// @Description Returns stock movement history for a warehouse
// @Tags stock-ledger
// @Produce json
// @Security BearerAuth
// @Param warehouse_id path string true "Warehouse ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stock-ledger/warehouse/{warehouse_id} [get]
func (h *StockLedgerHandler) GetWarehouseLedger(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("warehouse_id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	entries, total, err := h.ledgerUsecase.GetWarehouseLedger(c.Request.Context(), warehouseID, limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"ledger": entries,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// GetLedgerByDateRange godoc
// @Summary Get ledger by date range
// @Description Returns stock ledger entries between two dates
// @Tags stock-ledger
// @Produce json
// @Security BearerAuth
// @Param from query string true "Start date (RFC3339)"
// @Param to query string true "End date (RFC3339)"
// @Param warehouse_id query string false "Filter by warehouse"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stock-ledger/range [get]
func (h *StockLedgerHandler) GetLedgerByDateRange(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		utils.RespondBadRequest(c, "from and to dates are required")
		return
	}
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid from date; use RFC3339")
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid to date; use RFC3339")
		return
	}

	var warehouseID *uuid.UUID
	if raw := c.Query("warehouse_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			utils.RespondBadRequest(c, "Invalid warehouse ID")
			return
		}
		warehouseID = &id
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	entries, total, err := h.ledgerUsecase.GetLedgerByDateRange(c.Request.Context(), warehouseID, from, to, limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"ledger": entries,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}
