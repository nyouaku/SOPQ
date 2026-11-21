// ===== FILE: backend/internal/delivery/http/handler/product_change_request.go =====
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

type ProductChangeRequestHandler struct {
	uc usecase.ProductChangeRequestUsecase
}

func NewProductChangeRequestHandler(uc usecase.ProductChangeRequestUsecase) *ProductChangeRequestHandler {
	return &ProductChangeRequestHandler{uc: uc}
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	return middleware.GetUserIDFromContext(c)
}

// Submit godoc
// @Summary Ajukan perubahan master produk
// @Tags product-change-requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/product-change-requests [post]
func (h *ProductChangeRequestHandler) Submit(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		ChangeType string          `json:"change_type" validate:"required"`
		ProductID  *uuid.UUID      `json:"product_id"`
		Payload    json.RawMessage `json:"payload"`
		Reason     string          `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}
	out, err := h.uc.Submit(c.Request.Context(), usecase.SubmitChangeInput{
		ChangeType: req.ChangeType,
		ProductID:  req.ProductID,
		Payload:    req.Payload,
		Reason:     req.Reason,
		UserID:     userID,
	})
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SUBMIT_FAILED", err.Error())
		return
	}
	utils.RespondCreated(c, out)
}

// List godoc
// @Summary Daftar pengajuan perubahan produk
// @Tags product-change-requests
// @Produce json
// @Security BearerAuth
// @Router /api/v1/product-change-requests [get]
func (h *ProductChangeRequestHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	rows, total, err := h.uc.List(c.Request.Context(), c.Query("status"), limit, (page-1)*limit)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	utils.RespondOK(c, gin.H{
		"requests": rows,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// Approve godoc
// @Summary Setujui pengajuan perubahan produk
// @Tags product-change-requests
// @Produce json
// @Security BearerAuth
// @Router /api/v1/product-change-requests/:id/approve [post]
func (h *ProductChangeRequestHandler) Approve(c *gin.Context) {
	h.review(c, true)
}

// Reject godoc
// @Summary Tolak pengajuan perubahan produk
// @Tags product-change-requests
// @Produce json
// @Security BearerAuth
// @Router /api/v1/product-change-requests/:id/reject [post]
func (h *ProductChangeRequestHandler) Reject(c *gin.Context) {
	h.review(c, false)
}

func (h *ProductChangeRequestHandler) review(c *gin.Context, approve bool) {
	reviewerID, ok := currentUserID(c)
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid request ID")
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = c.ShouldBindJSON(&body)

	ip := c.ClientIP()
	if approve {
		err = h.uc.Approve(c.Request.Context(), id, reviewerID, body.Note, ip)
	} else {
		err = h.uc.Reject(c.Request.Context(), id, reviewerID, body.Note)
	}
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChangeRequestNotFound):
			utils.RespondError(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrChangeRequestNotPending):
			utils.RespondError(c, http.StatusConflict, "ALREADY_PROCESSED", err.Error())
		default:
			utils.RespondError(c, http.StatusBadRequest, "REVIEW_FAILED", err.Error())
		}
		return
	}
	utils.RespondOK(c, gin.H{"message": "Pengajuan diproses"})
}
