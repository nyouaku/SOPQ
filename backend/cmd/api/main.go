// ===== FILE: backend/cmd/api/main.go =====
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nyouaku/SOPQ/backend/internal/config"
	"github.com/nyouaku/SOPQ/backend/internal/delivery/http/route"
	"github.com/nyouaku/SOPQ/backend/internal/delivery/websocket"
	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/postgres"
	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/redis"
	"github.com/nyouaku/SOPQ/backend/pkg/logger"
)

// shutdownTimeout: batas tunggu request yang sedang berjalan sebelum proses dihentikan paksa.
// Harus lebih kecil dari stop_grace_period di docker-compose.
const shutdownTimeout = 30 * time.Second

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// Set logger level berdasarkan environment
	if cfg.App.Env == "production" {
		logger.SetLevel(logger.INFO)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.SetLevel(logger.DEBUG)
	}

	logger.Info("Starting Stock Opname API Server...")
	logger.Info("Environment: %s", cfg.App.Env)

	// Connect to PostgreSQL
	db, err := postgres.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Terapkan migrasi yang belum tercatat di schema_migrations.
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := db.RunMigrations(migrationsDir); err != nil {
		logger.Fatal("Migration gagal: %v", err)
	}

	// Connect to Redis (optional, tidak fatal jika gagal)
	var redisClient *redis.RedisClient
	redisClient, err = redis.NewRedisClient(&cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis: %v (continuing without Redis)", err)
		redisClient = nil
	} else {
		defer redisClient.Close()
	}

	// Setup Gin router
	router := gin.New()

	// Setup routes
	routerCfg := &route.RouterConfig{
		DB:                 db,
		RedisClient:        redisClient,
		Config:             cfg,
		CORSAllowedOrigins: cfg.Misc.CORSAllowedOrigins,
	}
	route.SetupRoutes(router, routerCfg)

	// http.Server (bukan router.Run) supaya bisa di-Shutdown dengan batas waktu.
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed to start: %v", err)
		}
	}()

	// Tunggu sinyal berhenti (docker stop mengirim SIGTERM).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Urutan: tutup koneksi WS dulu (klien akan reconnect), lalu hentikan listener
	// dan tunggu request aktif selesai. Defer di atas (db, redis) jalan setelah ini.
	websocket.ShutdownWS()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Warn("Graceful shutdown gagal, dipaksa berhenti: %v", err)
	}

	logger.Info("Server exited")
}

// ===== END FILE =====
