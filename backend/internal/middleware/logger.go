// ===== FILE: backend/internal/middleware/logger.go =====
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nyouaku/SOPQ/backend/pkg/logger"
)

// RequestIDHeader adalah header korelasi yang dikembalikan ke klien dan ikut di log.
const RequestIDHeader = "X-Request-ID"

// requestIDMaxLen membatasi nilai dari klien supaya tidak bisa memasukkan log raksasa.
const requestIDMaxLen = 64

// RequestID memakai X-Request-ID dari klien bila ada (dipotong), atau membuat ID baru.
// ID disimpan di context dan dikembalikan di response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" || len(id) > requestIDMaxLen {
			id = newRequestID()
		}
		c.Set("request_id", id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// Logger middleware untuk logging setiap request
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		reqID, _ := c.Get("request_id")

		if query != "" {
			path = path + "?" + query
		}

		recordRequest(method, c.FullPath(), statusCode, latency)

		logger.Info("%s | %3d | %13v | %15s | %s | rid=%v",
			method,
			statusCode,
			latency,
			clientIP,
			path,
			reqID,
		)
	}
}

// ===== END FILE =====
