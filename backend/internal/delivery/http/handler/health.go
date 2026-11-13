// ===== FILE: backend/internal/delivery/http/handler/health.go =====
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/postgres"
	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/redis"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db          *postgres.Database
	redisClient *redis.RedisClient
}

func NewHealthHandler(db *postgres.Database, redisClient *redis.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
	}
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// Health godoc
// @Summary Health check
// @Description Get health status of the application and its services
// @Tags health
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=HealthResponse}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	services := make(map[string]string)

	// Check database
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sqlDB, err := h.db.DB.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		services["database"] = "unhealthy"
	} else {
		services["database"] = "healthy"
	}

	// Check Redis
	if h.redisClient != nil {
		if err := h.redisClient.Client.Ping(ctx).Err(); err != nil {
			services["redis"] = "unhealthy"
		} else {
			services["redis"] = "healthy"
		}
	} else {
		services["redis"] = "not_configured"
	}

	// Determine overall status
	status := "healthy"
	for _, s := range services {
		if s == "unhealthy" {
			status = "unhealthy"
			break
		}
	}

	response := HealthResponse{
		Status:   status,
		Services: services,
	}

	if status == "unhealthy" {
		utils.RespondError(c, http.StatusServiceUnavailable, "UNHEALTHY", "Some services are unhealthy")
		return
	}

	utils.RespondOK(c, response)
}
// ===== END FILE =====
