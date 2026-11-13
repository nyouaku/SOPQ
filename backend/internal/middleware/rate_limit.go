// ===== FILE: backend/internal/middleware/rate_limit.go =====
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/redis"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

// incrWithTTL menaikkan counter dan memasang TTL dalam satu langkah atomik.
// INCR lalu EXPIRE terpisah punya celah: bila proses mati di antara keduanya,
// key tidak pernah kedaluwarsa dan IP itu terkunci selamanya.
var incrWithTTL = `
local c = redis.call('INCR', KEYS[1])
if c == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return c
`

// RateLimiter membatasi request per IP.
func RateLimiter(redisClient *redis.RedisClient, limit int, window time.Duration) gin.HandlerFunc {
	return RateLimiterByKey(redisClient, "ip", limit, window, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// LoginLimiter membatasi percobaan login per IP (limit) DAN per username
// (limit/2), supaya penyerang yang berpindah IP tetap tertahan oleh batas per akun.
func LoginLimiter(redisClient *redis.RedisClient, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Body dibaca untuk mengambil username, lalu dikembalikan utuh supaya
		// handler tetap bisa ShouldBindJSON.
		username := ""
		if body, err := io.ReadAll(io.LimitReader(c.Request.Body, 4096)); err == nil {
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			var probe struct {
				Username string `json:"username"`
			}
			if json.Unmarshal(body, &probe) == nil {
				username = probe.Username
			}
		}

		if !allow(redisClient, "login_ip:"+c.ClientIP(), limit, window) ||
			(username != "" && !allow(redisClient, "login_user:"+username, limit/2, window)) {
			utils.RespondError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Terlalu banyak percobaan login, coba lagi nanti")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimiterByKey membatasi request berdasarkan kunci yang dihitung keyFn.
func RateLimiterByKey(redisClient *redis.RedisClient, scope string, limit int, window time.Duration, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !allow(redisClient, scope+":"+keyFn(c), limit, window) {
			utils.RespondError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Terlalu banyak request, coba lagi nanti")
			c.Abort()
			return
		}
		c.Next()
	}
}

// allow mengembalikan true bila request masih di bawah batas.
// Redis tidak tersedia = fail open: rate limiting bukan kontrol keamanan
// utama, dan memblokir seluruh trafik saat Redis down lebih merugikan.
func allow(redisClient *redis.RedisClient, key string, limit int, window time.Duration) bool {
	if redisClient == nil {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := redisClient.Client.Eval(ctx, incrWithTTL,
		[]string{fmt.Sprintf("rate_limit:%s", key)}, window.Milliseconds()).Int64()
	if err != nil {
		return true
	}
	return count <= int64(limit)
}

// ===== END FILE =====
