// ===== FILE: backend/internal/delivery/http/handler/inventory.go =====
package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
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

// CancelMovementRequest — body pembatalan dokumen pergerakan.
type CancelMovementRequest struct {
	Reason string `json:"reason" validate:"required,min=3,max=500"`
}

// CancelMovement godoc
// @Summary Batalkan dokumen pergerakan stok
// @Description Membalik seluruh efek dokumen (stok & ledger) lalu menandainya 'cancelled'.
// @Description Dokumen TIDAK dihapus agar audit trail utuh. Hanya dokumen 'posted' bisa dibatalkan.
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Movement ID"
// @Param body body CancelMovementRequest true "Alasan pembatalan"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/{id}/cancel [post]
func (h *InventoryHandler) CancelMovement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid movement ID")
		return
	}
	var req CancelMovementRequest
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

	err = h.inventoryUsecase.CancelMovement(c.Request.Context(), id, uid, req.Reason, c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMovementNotFound):
			utils.RespondNotFound(c, err.Error())
		case errors.Is(err, domain.ErrMovementAlreadyCancelled):
			utils.RespondConflict(c, err.Error())
		case errors.Is(err, domain.ErrMovementNotCancellable):
			utils.RespondError(c, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		default:
			utils.RespondBadRequest(c, err.Error())
		}
		return
	}

	utils.RespondOK(c, gin.H{"status": domain.MovementStatusCancelled})
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

// gridEntryRequest satu sel isian grid.
type gridEntryRequest struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Date      string  `json:"date" validate:"required"` // YYYY-MM-DD
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	UnitCost  float64 `json:"unit_cost"`
}

// PostGridRequest — DTO posting seluruh isian grid sekaligus.
type PostGridRequest struct {
	WarehouseID   string            `json:"warehouse_id" validate:"required,uuid"`
	MovementType  string            `json:"movement_type" validate:"required,oneof=in out"`
	SubType       string            `json:"sub_type" validate:"required"`
	ReferenceCode string            `json:"reference_code"`
	Entries       []gridEntryRequest `json:"entries" validate:"required,min=1,dive"`
}

// PostGrid godoc
// @Summary Post inventory grid (bulk, per tanggal)
// @Description Simpan seluruh isian grid (produk × tanggal) sekaligus. Item dikelompokkan
// @Description per tanggal -> satu dokumen per tanggal, tiap dokumen atomic.
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PostGridRequest true "Grid data"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/grid [post]
func (h *InventoryHandler) PostGrid(c *gin.Context) {
	var req PostGridRequest
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

	entries := make([]usecase.GridEntryInput, 0, len(req.Entries))
	for _, e := range req.Entries {
		pid, perr := uuid.Parse(e.ProductID)
		if perr != nil {
			utils.RespondBadRequest(c, "Invalid product ID")
			return
		}
		d, derr := time.Parse("2006-01-02", e.Date)
		if derr != nil {
			utils.RespondBadRequest(c, "Tanggal harus format YYYY-MM-DD")
			return
		}
		entries = append(entries, usecase.GridEntryInput{
			ProductID: pid,
			Date:      d,
			Quantity:  e.Quantity,
			UnitCost:  e.UnitCost,
		})
	}

	result, err := h.inventoryUsecase.PostGrid(c.Request.Context(), usecase.PostGridInput{
		WarehouseID:   whID,
		MovementType:  req.MovementType,
		SubType:       req.SubType,
		ReferenceCode: req.ReferenceCode,
		UserID:        uid,
		IPAddress:     c.ClientIP(),
		Entries:       entries,
	})
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}
	utils.RespondOK(c, result)
}

// GetGrid godoc
// @Summary Get inventory grid (pre-fill per produk × tanggal)
// @Description Agregat isian grid per produk × tanggal untuk satu gudang + bulan + tipe.
// @Tags inventory
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string true "Warehouse ID"
// @Param month query string true "Month (YYYY-MM)"
// @Param movement_type query string true "in | out"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/grid [get]
func (h *InventoryHandler) GetGrid(c *gin.Context) {
	whRaw := c.Query("warehouse_id")
	if whRaw == "" {
		utils.RespondBadRequest(c, "warehouse_id wajib")
		return
	}
	whID, err := uuid.Parse(whRaw)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}
	monthRaw := c.Query("month")
	if monthRaw == "" {
		utils.RespondBadRequest(c, "month wajib (YYYY-MM)")
		return
	}
	month, err := time.Parse("2006-01", monthRaw)
	if err != nil {
		utils.RespondBadRequest(c, "month harus format YYYY-MM")
		return
	}
	mt := c.Query("movement_type")
	if mt != "in" && mt != "out" {
		utils.RespondBadRequest(c, "movement_type harus 'in' atau 'out'")
		return
	}

	cells, err := h.inventoryUsecase.GetGrid(c.Request.Context(), usecase.GridQuery{
		WarehouseID:  whID,
		Month:        month,
		MovementType: mt,
	})
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	utils.RespondOK(c, gin.H{"cells": cells})
}

// GetSummary godoc
// @Summary Get inventory dashboard KPI (server-side)
// @Description KPI dashboard: total masuk/keluar/transfer, nilai, top keluar, stok menipis.
// @Tags inventory
// @Produce json
// @Security BearerAuth
// @Param warehouse_id query string true "Warehouse ID"
// @Param month query string true "Month (YYYY-MM)"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/summary [get]
func (h *InventoryHandler) GetSummary(c *gin.Context) {
	whRaw := c.Query("warehouse_id")
	if whRaw == "" {
		utils.RespondBadRequest(c, "warehouse_id wajib")
		return
	}
	whID, err := uuid.Parse(whRaw)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}
	monthRaw := c.DefaultQuery("month", time.Now().Format("2006-01"))
	month, err := time.Parse("2006-01", monthRaw)
	if err != nil {
		utils.RespondBadRequest(c, "month harus format YYYY-MM")
		return
	}

	summary, err := h.inventoryUsecase.GetSummary(c.Request.Context(), usecase.SummaryQuery{
		WarehouseID: whID,
		Month:       month,
	})
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	utils.RespondOK(c, summary)
}

// CopyGridRequest — DTO salin isian grid bulan sumber ke bulan tujuan.
type CopyGridRequest struct {
	WarehouseID   string `json:"warehouse_id" validate:"required,uuid"`
	SrcMonth      string `json:"src_month" validate:"required"` // YYYY-MM
	DstMonth      string `json:"dst_month" validate:"required"` // YYYY-MM
	MovementType  string `json:"movement_type" validate:"required,oneof=in out"`
	SubType       string `json:"sub_type" validate:"required"`
}

// CopyGrid godoc
// @Summary Copy inventory grid dari bulan sumber ke bulan tujuan
// @Description Salin seluruh isian grid (produk & hari sama) dari bulan sumber ke bulan tujuan.
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CopyGridRequest true "Copy grid data"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/inventory/grid/copy [post]
func (h *InventoryHandler) CopyGrid(c *gin.Context) {
	var req CopyGridRequest
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
	srcMonth, err := time.Parse("2006-01", req.SrcMonth)
	if err != nil {
		utils.RespondBadRequest(c, "src_month harus format YYYY-MM")
		return
	}
	dstMonth, err := time.Parse("2006-01", req.DstMonth)
	if err != nil {
		utils.RespondBadRequest(c, "dst_month harus format YYYY-MM")
		return
	}

	result, err := h.inventoryUsecase.CopyGrid(c.Request.Context(), usecase.CopyGridInput{
		WarehouseID:  whID,
		SrcMonth:     srcMonth,
		DstMonth:     dstMonth,
		MovementType: req.MovementType,
		SubType:      req.SubType,
		UserID:       uid,
		IPAddress:    c.ClientIP(),
	})
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}
	utils.RespondOK(c, result)
}

// ExportGrid godoc
// @Summary Export inventory grid ke Excel
// @Description Unduh grid harian (produk × tanggal) sebagai file Excel sesuai kebiasaan kerja.
// @Tags inventory
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param warehouse_id query string true "Warehouse ID"
// @Param month query string true "Month (YYYY-MM)"
// @Param movement_type query string true "in | out"
// @Success 200 {file} file
// @Router /api/v1/inventory/grid/export [get]
func (h *InventoryHandler) ExportGrid(c *gin.Context) {
	whRaw := c.Query("warehouse_id")
	if whRaw == "" {
		utils.RespondBadRequest(c, "warehouse_id wajib")
		return
	}
	whID, err := uuid.Parse(whRaw)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}
	monthRaw := c.Query("month")
	if monthRaw == "" {
		utils.RespondBadRequest(c, "month wajib (YYYY-MM)")
		return
	}
	month, err := time.Parse("2006-01", monthRaw)
	if err != nil {
		utils.RespondBadRequest(c, "month harus format YYYY-MM")
		return
	}
	mt := c.Query("movement_type")
	if mt != "in" && mt != "out" {
		utils.RespondBadRequest(c, "movement_type harus 'in' atau 'out'")
		return
	}

	file, filename, err := h.inventoryUsecase.ExportGridToExcel(c.Request.Context(), usecase.GridQuery{
		WarehouseID:  whID,
		Month:        month,
		MovementType: mt,
	})
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	if err := file.Write(c.Writer); err != nil {
		utils.RespondInternalError(c, err.Error())
	}
}

// ===== END FILE =====
