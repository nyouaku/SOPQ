// ===== FILE: backend/internal/middleware/cors.go =====
package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS mengizinkan origin yang terdaftar di CORS_ALLOWED_ORIGINS (dipisah koma).
//
// Aturan penting: `Access-Control-Allow-Origin: *` TIDAK BOLEH digabung dengan
// `Access-Control-Allow-Credentials: true`. Kombinasi itu ditolak browser, dan
// bila dipaksakan akan membuat situs mana pun bisa membaca respons ber-kredensial.
// Jadi wildcard => tanpa credentials; origin eksplisit => boleh credentials.
func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := parseOrigins(allowedOrigins)
	wildcard := false
	for _, o := range origins {
		if o == "*" {
			wildcard = true
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowed := wildcard || (origin != "" && slices.Contains(origins, origin))

		if allowed {
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Add("Vary", "Origin")
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
			if !wildcard {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Max-Age", "600")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func parseOrigins(raw string) []string {
	var out []string
	for o := range strings.SplitSeq(raw, ",") {
		if o = strings.TrimSpace(strings.TrimSuffix(o, "/")); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// ===== END FILE =====
