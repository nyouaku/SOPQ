// ===== FILE: backend/internal/middleware/rbac.go =====
package middleware

import (
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// RequireRole middleware untuk memastikan user memiliki role tertentu
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetUserFromContext(c)
		if !ok {
			utils.RespondUnauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Cek apakah role user ada di daftar allowed roles
		allowed := false
		for _, role := range allowedRoles {
			if claims.RoleName == role {
				allowed = true
				break
			}
		}

		if !allowed {
			utils.RespondForbidden(c, "Anda tidak memiliki akses ke resource ini")
			c.Abort()
			return
		}

		c.Next()
	}
}
// ===== END FILE =====
