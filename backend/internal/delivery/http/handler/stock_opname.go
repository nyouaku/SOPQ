// ===== FILE: backend/internal/delivery/http/handler/stock_opname.go =====
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/nyouaku/SOPQ/backend/pkg/validator"
)

type StockOpnameHandler struct {
	stockOpnameUsecase usecase.StockOpnameUsecase
}

func NewStockOpnameHandler(stockOpnameUsecase usecase.StockOpnameUsecase) *StockOpnameHandler {
	return &StockOpnameHandler{
		stockOpnameUsecase: stockOpnameUsecase,
	}
}

// DTOs
type StartSessionRequest struct {
	WarehouseID uuid.UUID  `json:"warehouse_id" validate:"required"`
	LocationID  *uuid.UUID `json:"location_id"`
}

type SubmitScanRequest struct {
	ProductCode   string     `json:"product_code" validate:"required"`
	StockPhysical int        `json:"stock_physical" validate:"required,min=0"`
	InputMethod   string     `json:"input_method" validate:"required,oneof=manual voice camera upload"`
	Note          string     `json:"note"`
	DeviceID      *uuid.UUID `json:"device_id"`
	// IdempotencyKey opsional: UUID per scan dari klien mobile. Retry dengan
	// kunci sama mengembalikan baris yang sudah ada, bukan membuat duplikat.
	IdempotencyKey string `json:"idempotency_key" validate:"omitempty,uuid4|uuid"`
}

// StartSession godoc
// @Summary Start stock opname session
// @Description Create new stock opname session (supervisor/staff)
// @Tags stock-opname
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body StartSessionRequest true "Session data"
// @Success 201 {object} utils.SuccessResponse{data=domain.StockOpname}
// @Router /api/v1/stock-opname/sessions [post]
func (h *StockOpnameHandler) StartSession(c *gin.Context) {
	var req StartSessionRequest
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
		utils.RespondUnauthorized(c, "User not authenticated")
		return
	}

	input := usecase.StartSessionInput{
		WarehouseID: req.WarehouseID,
		LocationID:  req.LocationID,
		StartedBy:   userID,
		IPAddress:   c.ClientIP(),
	}

	session, err := h.stockOpnameUsecase.StartSession(c.Request.Context(), input)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "START_SESSION_FAILED", err.Error())
		return
	}

	utils.RespondCreated(c, session)
}

// SubmitScan godoc
// @Summary Submit scan result
// @Description Submit product scan (staff)
// @Tags stock-opname
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param request body SubmitScanRequest true "Scan data"
// @Success 201 {object} utils.SuccessResponse{data=domain.StockOpnameDetail}
// @Router /api/v1/stock-opname/sessions/{id}/scan [post]
func (h *StockOpnameHandler) SubmitScan(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid session ID")
		return
	}

	var req SubmitScanRequest
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
		utils.RespondUnauthorized(c, "User not authenticated")
		return
	}

	input := usecase.SubmitScanInput{
		SessionID:      sessionID,
		ProductCode:    req.ProductCode,
		StockPhysical:  req.StockPhysical,
		InputMethod:    req.InputMethod,
		Note:           req.Note,
		ScannedBy:      userID,
		DeviceID:       req.DeviceID,
		IPAddress:      c.ClientIP(),
		IdempotencyKey: req.IdempotencyKey,
	}

	detail, err := h.stockOpnameUsecase.SubmitScan(c.Request.Context(), input)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SUBMIT_SCAN_FAILED", err.Error())
		return
	}

	utils.RespondCreated(c, detail)
}

// GetSession godoc
// @Summary Get session details
// @Description Get stock opname session info
// @Tags stock-opname
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} utils.SuccessResponse{data=domain.StockOpname}
// @Router /api/v1/stock-opname/sessions/{id} [get]
func (h *StockOpnameHandler) GetSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid session ID")
		return
	}

	session, err := h.stockOpnameUsecase.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, session)
}

// CompleteSession godoc
// @Summary Complete session
// @Description Mark session as completed (supervisor/staff)
// @Tags stock-opname
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stock-opname/sessions/{id}/complete [post]
func (h *StockOpnameHandler) CompleteSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid session ID")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.RespondUnauthorized(c, "User not authenticated")
		return
	}

	if err := h.stockOpnameUsecase.CompleteSession(c.Request.Context(), sessionID, userID); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "COMPLETE_FAILED", err.Error())
		return
	}

	utils.RespondOK(c, gin.H{"message": "Session completed successfully"})
}

// ApproveSession godoc
// @Summary Approve session
// @Description Approve completed session (supervisor only)
// @Tags stock-opname
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stock-opname/sessions/{id}/approve [post]
func (h *StockOpnameHandler) ApproveSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid session ID")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.RespondUnauthorized(c, "User not authenticated")
		return
	}

	if err := h.stockOpnameUsecase.ApproveSession(c.Request.Context(), sessionID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			utils.RespondNotFound(c, "Session not found")
		case errors.Is(err, domain.ErrSessionAlreadyApproved):
			utils.RespondConflict(c, "Session is already approved")
		case errors.Is(err, domain.ErrSessionNotApprovable):
			utils.RespondError(c, http.StatusBadRequest, "APPROVE_FAILED", "Session must be completed before approval")
		default:
			utils.RespondError(c, http.StatusInternalServerError, "APPROVE_FAILED", "Failed to approve session")
		}
		return
	}

	utils.RespondOK(c, gin.H{"message": "Session approved successfully"})
}

// GetSessionDetails godoc
// @Summary Get session details (scanned items)
// @Description Get list of scanned items in a session
// @Tags stock-opname
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stock-opname/sessions/{id}/details [get]
func (h *StockOpnameHandler) GetSessionDetails(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid session ID")
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

	details, total, err := h.stockOpnameUsecase.GetSessionDetails(c.Request.Context(), sessionID, limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"details": details,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// ListSessions godoc
// @Summary List stock opname sessions
// @Description List sessions with filters
// @Tags stock-opname
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param warehouse_id query string false "Filter by warehouse"
// @Param status query string false "Filter by status"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/stock-opname/sessions [get]
func (h *StockOpnameHandler) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	filter := repository.StockOpnameFilter{
		Status: c.Query("status"),
		Limit:  limit,
		Offset: offset,
	}

	if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
		id, err := uuid.Parse(warehouseID)
		if err == nil {
			filter.WarehouseID = &id
		}
	}

	sessions, total, err := h.stockOpnameUsecase.ListSessions(c.Request.Context(), filter)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// ===== END FILE =====
