// ===== FILE: backend/pkg/utils/response.go =====
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondSuccess mengirimkan response sukses dengan format standar
func RespondSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// RespondError mengirimkan response error dengan format standar
func RespondError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// RespondOK shorthand untuk 200 OK
func RespondOK(c *gin.Context, data interface{}) {
	RespondSuccess(c, http.StatusOK, data, "")
}

// RespondCreated shorthand untuk 201 Created
func RespondCreated(c *gin.Context, data interface{}) {
	RespondSuccess(c, http.StatusCreated, data, "")
}

// RespondBadRequest shorthand untuk 400 Bad Request
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// RespondUnauthorized shorthand untuk 401 Unauthorized
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// RespondForbidden shorthand untuk 403 Forbidden
func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, "FORBIDDEN", message)
}

// RespondNotFound shorthand untuk 404 Not Found
func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// RespondConflict shorthand untuk 409 Conflict
func RespondConflict(c *gin.Context, message string) {
	RespondError(c, http.StatusConflict, "CONFLICT", message)
}

// RespondInternalError shorthand untuk 500 Internal Server Error
func RespondInternalError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// RespondValidationError shorthand untuk 422 Unprocessable Entity
func RespondValidationError(c *gin.Context, message string) {
	RespondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message)
}

// ===== END FILE =====
