// ===== FILE: backend/internal/delivery/http/handler/auth.go =====
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/logger"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/nyouaku/SOPQ/backend/pkg/validator"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

// DTOs
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required,min=3"`
	RoleID   int    `json:"role_id" validate:"required,min=1,max=3"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Register godoc
// @Summary Register new user
// @Description Create new user account (admin only)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register data"
// @Success 201 {object} utils.SuccessResponse{data=domain.User}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	input := usecase.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		FullName: req.FullName,
		RoleID:   req.RoleID,
	}

	user, err := h.authUsecase.Register(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUsernameTaken):
			utils.RespondConflict(c, err.Error())
		case errors.Is(err, domain.ErrWeakPassword), errors.Is(err, domain.ErrRoleNotFound):
			utils.RespondBadRequest(c, err.Error())
		default:
			logger.Error("Register failed: %v", err)
			utils.RespondInternalError(c, "Gagal membuat user")
		}
		return
	}

	utils.RespondCreated(c, user)
}

// Login godoc
// @Summary Login
// @Description Login with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} utils.SuccessResponse{data=usecase.LoginResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	input := usecase.LoginInput{
		Username: req.Username,
		Password: req.Password,
		DeviceInfo: usecase.DeviceInfo{
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		},
	}

	response, err := h.authUsecase.Login(c.Request.Context(), input)
	if err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	utils.RespondOK(c, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange refresh token for new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} utils.SuccessResponse{data=usecase.LoginResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	deviceInfo := usecase.DeviceInfo{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	response, err := h.authUsecase.RefreshToken(c.Request.Context(), req.RefreshToken, deviceInfo)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRefreshTokenReused):
			// Token rotation mendeteksi pemakaian ulang: sesi ini dicurigai dicuri.
			logger.Warn("Refresh token reused; all sessions of the user revoked")
			utils.RespondUnauthorized(c, "Refresh token has already been used")
		case errors.Is(err, domain.ErrRefreshTokenInvalid), errors.Is(err, domain.ErrUserInactive):
			utils.RespondUnauthorized(c, "Refresh token is invalid or expired")
		default:
			logger.Error("Refresh token failed: %v", err)
			utils.RespondError(c, http.StatusInternalServerError, "REFRESH_FAILED", "Failed to refresh token")
		}
		return
	}

	utils.RespondOK(c, response)
}

// Logout godoc
// @Summary Logout
// @Description Revoke refresh token session
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.RespondUnauthorized(c, "User not authenticated")
		return
	}

	// Revoke all sessions for this user
	if err := h.authUsecase.Logout(c.Request.Context(), userID); err != nil {
		logger.Error("Failed to logout user %s: %v", userID, err)
		utils.RespondError(c, http.StatusInternalServerError, "LOGOUT_FAILED", "Failed to revoke sessions")
		return
	}

	utils.RespondOK(c, gin.H{
		"message": "Logout berhasil, semua session telah di-revoke",
		"user_id": userID,
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=domain.User}
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.RespondUnauthorized(c, "User not authenticated")
		return
	}

	user, err := h.authUsecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "FAILED", err.Error())
		return
	}

	utils.RespondOK(c, user)
}

// ListUsers godoc
// @Summary List semua user
// @Description Daftar user untuk manajemen staff (admin only)
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Router /api/v1/auth/users [get]
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.authUsecase.ListUsers(c.Request.Context())
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	utils.RespondOK(c, gin.H{"users": users})
}

// ListRoles godoc
// @Summary List role
// @Description Daftar role yang tersedia (admin/supervisor/staff_gudang) untuk form register staff
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Router /api/v1/auth/roles [get]
func (h *AuthHandler) ListRoles(c *gin.Context) {
	roles, err := h.authUsecase.ListRoles(c.Request.Context())
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	utils.RespondOK(c, gin.H{"roles": roles})
}

// ===== END FILE =====
