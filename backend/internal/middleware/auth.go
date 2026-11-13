// ===== FILE: backend/internal/middleware/auth.go =====
package middleware

import (
	"strings"

	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
)

// AuthMiddleware memvalidasi JWT access token dan inject user claims ke context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			utils.RespondUnauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			utils.RespondUnauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)

		claims, err := utils.ParseAccessToken(token, jwtSecret)
		if err != nil {
			utils.RespondUnauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Inject user claims ke context
		c.Set(UserContextKey, claims)
		c.Next()
	}
}

// GetUserFromContext mengambil user claims dari context
func GetUserFromContext(c *gin.Context) (*utils.JWTClaims, bool) {
	value, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}
	claims, ok := value.(*utils.JWTClaims)
	return claims, ok
}

// GetUserIDFromContext mengambil user_id dari context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return uuid.Nil, false
	}
	return claims.UserID, true
}
// ===== END FILE =====
