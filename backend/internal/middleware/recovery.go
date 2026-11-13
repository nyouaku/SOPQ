// ===== FILE: backend/internal/middleware/recovery.go =====
package middleware

import (
	"net/http"

	"github.com/nyouaku/SOPQ/backend/pkg/logger"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Recovery middleware untuk menangkap panic dan mengembalikan response error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered: %v", err)
				utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server")
				c.Abort()
			}
		}()
		c.Next()
	}
}
// ===== END FILE =====
