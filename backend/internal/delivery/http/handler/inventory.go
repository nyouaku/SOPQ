// ===== FILE: backend/internal/delivery/http/handler/inventory.go =====
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/nyouaku/SOPQ/backend/pkg/validator"
)

type InventoryHandler struct {
	inventoryUsecase usecase.InventoryUsecase
}

func NewInventoryHandler(inventoryUsecase usecase.InventoryUsecase) *InventoryHandler {
	return &InventoryHandler{inventoryUsecase: inventoryUsecase}
}

// movementItemRequest satu baris produk.
type movementItemRequest struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	UnitCost  float64 `json:"unit_cost"`
	Note      string  `json:"note"`
}

func (h *InventoryHandler) userID(c *gin.Context) (uuid.UUID, bool) {
	return middleware.GetUserIDFromContext(c)
}

func (h *InventoryHandler) toItems(raw []movementItemRequest) ([]usecase.MovementItemInput, error) {
	items := make([]usecase.MovementItemInput, 0, len(raw))
	for _, r := range raw {
		pid, err := uuid.Parse(r.ProductID)
		if err != nil {
			return nil, err
		}
		items = append(items, usecase.MovementItemInput{
			ProductID: pid,
			Quantity:  r.Quantity,
			UnitCost:  r.UnitCost,
			Note:      r.Note,
		})
	}
	return items, nil
}

// PostIn godoc
// @Summary Post stock IN (barang masuk)
// @Description Catat barang masuk ke gudang (pembelian, retur pelanggan). Naikkan stok + tulis ledger.
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PostMovementRequest true "Stock IN data"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/in [post]
func (h *InventoryHandler) PostIn(c *gin.Context) {
	var req PostMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}
	uid, ok := h.userID(c)
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	whID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}
	items, err := h.toItems(req.Items)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}

	movement, err := h.inventoryUsecase.PostIn(c.Request.Context(), usecase.PostMovementInput{
		WarehouseID:   whID,
		SubType:       req.SubType,
		ReferenceCode: req.ReferenceCode,
		Note:          req.Note,
		UserID:        uid,
		IPAddress:     c.ClientIP(),
		Items:         items,
	})
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}
	utils.RespondOK(c, movement)
}

// PostOut godoc
// @Summary Post stock OUT (barang keluar)
// @Description Catat barang keluar dari gudang (penjualan, rusak, hilang, expired). Turunkan stok + tulis ledger.
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PostMovementRequest true "Stock OUT data"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/out [post]
func (h *InventoryHandler) PostOut(c *gin.Context) {
	var req PostMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}
	uid, ok := h.userID(c)
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	whID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}
	items, err := h.toItems(req.Items)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}

	movement, err := h.inventoryUsecase.PostOut(c.Request.Context(), usecase.PostMovementInput{
		WarehouseID:   whID,
		SubType:       req.SubType,
		ReferenceCode: req.ReferenceCode,
		Note:          req.Note,
		UserID:        uid,
		IPAddress:     c.ClientIP(),
		Items:         items,
	})
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}
	utils.RespondOK(c, movement)
}

// PostTransfer godoc
// @Summary Post stock transfer antar gudang
// @Description Pindahkan stok antar dua gudang. Gudang asal turun, tujuan naik, 2 baris ledger.
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PostTransferRequest true "Transfer data"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/transfer [post]
func (h *InventoryHandler) PostTransfer(c *gin.Context) {
	var req PostTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}
	uid, ok := h.userID(c)
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	fromID, err := uuid.Parse(req.FromWarehouseID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid from warehouse ID")
		return
	}
	toID, err := uuid.Parse(req.ToWarehouseID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid to warehouse ID")
		return
	}
	items, err := h.toItems(req.Items)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}

	movement, err := h.inventoryUsecase.PostTransfer(c.Request.Context(), usecase.PostTransferInput{
		FromWarehouseID: fromID,
		ToWarehouseID:   toID,
		ReferenceCode:   req.ReferenceCode,
		Note:            req.Note,
		UserID:          uid,
		IPAddress:       c.ClientIP(),
		Items:           items,
	})
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}
	utils.RespondOK(c, movement)
}

// GetMovement godoc
// @Summary Get inventory movement detail
// @Description Ambil satu dokumen pergerakan stok beserta item-nya.
// @Tags inventory
// @Produce json
// @Security BearerAuth
// @Param id path string true "Movement ID"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/{id} [get]
func (h *InventoryHandler) GetMovement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid movement ID")
		return
	}
	movement, err := h.inventoryUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	if movement == nil {
		utils.RespondNotFound(c, "Pergerakan stok tidak ditemukan")
		return
	}
	utils.RespondOK(c, movement)
}

// ListMovements godoc
// @Summary List inventory movements
// @Description Daftar dokumen pergerakan stok dengan filter gudang/tipe/rentang tanggal.
// @Tags inventory
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string false "Filter by warehouse"
// @Param movement_type query string false "in | out | transfer"
// @Param from query string false "YYYY-MM-DD"
// @Param to query string false "YYYY-MM-DD"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(25)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory [get]
func (h *InventoryHandler) ListMovements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}

	filter := buildInventoryFilter(c)
	filter.Limit = limit
	filter.Offset = (page - 1) * limit
	movements, total, err := h.inventoryUsecase.List(c.Request.Context(), filter)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	utils.RespondOK(c, gin.H{
		"data":  movements,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// buildInventoryFilter membangun filter list dari query string.
func buildInventoryFilter(c *gin.Context) repository.InventoryFilter {
	var f repository.InventoryFilter
	if raw := c.Query("warehouse_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			f.WarehouseID = &id
		}
	}
	f.MovementType = c.Query("movement_type")
	if raw := c.Query("from"); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			f.From = &t
		}
	}
	if raw := c.Query("to"); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			end := t.Add(24 * time.Hour) // batas atas inklusif
			f.To = &end
		}
	}
	return f
}

// PostMovementRequest — DTO IN/OUT.
type PostMovementRequest struct {
	WarehouseID   string                `json:"warehouse_id" validate:"required,uuid"`
	SubType       string                `json:"sub_type" validate:"required"`
	ReferenceCode string                `json:"reference_code"`
	Note          string                `json:"note"`
	Items         []movementItemRequest `json:"items" validate:"required,min=1,dive"`
}

// PostTransferRequest — DTO transfer antar gudang.
type PostTransferRequest struct {
	FromWarehouseID string                `json:"from_warehouse_id" validate:"required,uuid"`
	ToWarehouseID   string                `json:"to_warehouse_id" validate:"required,uuid"`
	ReferenceCode   string                `json:"reference_code"`
	Note            string                `json:"note"`
	Items           []movementItemRequest `json:"items" validate:"required,min=1,dive"`
}

// ===== END FILE =====
