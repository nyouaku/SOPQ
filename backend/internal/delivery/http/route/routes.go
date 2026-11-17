package route

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/config"
	"github.com/nyouaku/SOPQ/backend/internal/delivery/http/handler"
	"github.com/nyouaku/SOPQ/backend/internal/delivery/websocket"
	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/postgres"
	"github.com/nyouaku/SOPQ/backend/internal/infrastructure/redis"
	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
)

// canAccessSession membatasi langganan WebSocket: admin & supervisor boleh
// melihat sesi mana pun, staf hanya sesi yang ia mulai atau yang warehouse-nya
// sama dengan sesi yang ia buat.
func canAccessSession(uc usecase.StockOpnameUsecase) func(context.Context, uuid.UUID, string, uuid.UUID) (bool, error) {
	return func(ctx context.Context, userID uuid.UUID, roleName string, sessionID uuid.UUID) (bool, error) {
		if roleName == "admin" || roleName == "supervisor" {
			return true, nil
		}
		session, err := uc.GetSessionByID(ctx, sessionID)
		if err != nil {
			return false, err
		}
		if session == nil {
			return false, nil
		}
		return session.StartedBy == userID, nil
	}
}

type RouterConfig struct {
	DB                 *postgres.Database
	RedisClient        *redis.RedisClient
	Config             *config.Config
	CORSAllowedOrigins string
}

// SetupRoutes mengatur semua routes aplikasi
func SetupRoutes(router *gin.Engine, cfg *RouterConfig) {
	// Middleware global
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Health endpoint (tanpa prefix)
	healthHandler := handler.NewHealthHandler(cfg.DB, cfg.RedisClient)
	router.GET("/health", healthHandler.Health)
	// Metrik Prometheus. Jangan diekspos ke publik: nginx sebaiknya memblokir /metrics.
	router.GET("/metrics", middleware.MetricsHandler())

	// Initialize repositories
	userRepo := repository.NewUserRepository(cfg.DB.DB)
	sessionRepo := repository.NewSessionRepository(cfg.DB.DB)
	categoryRepo := repository.NewCategoryRepository(cfg.DB.DB)
	supplierRepo := repository.NewSupplierRepository(cfg.DB.DB)
	warehouseRepo := repository.NewWarehouseRepository(cfg.DB.DB)
	locationRepo := repository.NewLocationRepository(cfg.DB.DB)
	boxRepo := repository.NewBoxRepository(cfg.DB.DB)
	productRepo := repository.NewProductRepository(cfg.DB.DB)
	stockOpnameRepo := repository.NewStockOpnameRepository(cfg.DB.DB)
	stockOpnameDetailRepo := repository.NewStockOpnameDetailRepository(cfg.DB.DB)
	stockRepo := repository.NewStockRepository(cfg.DB.DB)
	activityLogRepo := repository.NewActivityLogRepository(cfg.DB.DB)
	notificationRepo := repository.NewNotificationRepository(cfg.DB.DB)
	stockLedgerRepo := repository.NewStockLedgerRepository(cfg.DB.DB)
	inventoryRepo := repository.NewInventoryRepository(cfg.DB.DB)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, activityLogRepo, cfg.Config)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	supplierUsecase := usecase.NewSupplierUsecase(supplierRepo)
	warehouseUsecase := usecase.NewWarehouseUsecase(warehouseRepo, locationRepo, boxRepo)
	productUsecase := usecase.NewProductUsecase(productRepo)
	stockOpnameUsecase := usecase.NewStockOpnameUsecase(stockOpnameRepo, stockOpnameDetailRepo, stockRepo, productRepo, activityLogRepo, websocket.NewPublisher())
	// Suntikkan ledger repo agar adjustment & approve SO ikut mencatat stock_ledger.
	stockOpnameUsecase.SetLedgerRepo(stockLedgerRepo)
	reportUsecase := usecase.NewReportUsecase(stockOpnameDetailRepo, activityLogRepo, repository.NewReportRepository(cfg.DB.DB))
	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	dashboardUsecase := usecase.NewDashboardUsecase(repository.NewDashboardRepository(cfg.DB.DB), activityLogRepo)
	exportUsecase := usecase.NewExportUsecase(stockOpnameDetailRepo, productRepo, repository.NewReportRepository(cfg.DB.DB))
	importUsecase := usecase.NewImportUsecase(productRepo, categoryRepo, stockRepo, warehouseRepo)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, activityLogRepo)
	stockLedgerUsecase := usecase.NewStockLedgerUsecase(stockLedgerRepo, stockRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	supplierHandler := handler.NewSupplierHandler(supplierUsecase)
	warehouseHandler := handler.NewWarehouseHandler(warehouseUsecase)
	productHandler := handler.NewProductHandler(productUsecase)
	productHandler.SetStockRepos(stockRepo, stockLedgerRepo)
	stockOpnameHandler := handler.NewStockOpnameHandler(stockOpnameUsecase)
	reportHandler := handler.NewReportHandler(reportUsecase, notificationUsecase)
	dashboardHandler := handler.NewDashboardHandler(dashboardUsecase)
	exportHandler := handler.NewExportHandler(exportUsecase)

	snapshotHandler := handler.NewSnapshotHandler(cfg.DB.DB)
	importHandler := handler.NewImportHandler(importUsecase)
	stockLedgerHandler := handler.NewStockLedgerHandler(stockLedgerUsecase)
	reconciliationHandler := handler.NewReconciliationHandler(stockOpnameUsecase)
	stockAdjustmentHandler := handler.NewStockAdjustmentHandler(stockOpnameUsecase)
	inventoryHandler := handler.NewInventoryHandler(inventoryUsecase)
	backupHandler := handler.NewBackupHandler(cfg.Config, cfg.DB.DB)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			// SEMENTARA unlimited: rate limit login dimatikan untuk debug login (lihat TODO di bawah).
			// TODO: kembalikan ke LoginLimiter(cfg.RedisClient, 10, 15*time.Minute) dan
			// RateLimiterByKey(..., "refresh_ip", 30, 15*time.Minute, ...) setelah login beres.
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
			{
				authProtected.GET("/me", authHandler.GetProfile)
				authProtected.POST("/logout", authHandler.Logout)
			}

			// Admin only
			authAdmin := auth.Group("")
			authAdmin.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
			authAdmin.Use(middleware.RequireRole("admin"))
			{
				authAdmin.POST("/register", authHandler.Register)
				authAdmin.GET("/users", authHandler.ListUsers)
				authAdmin.GET("/roles", authHandler.ListRoles)
				authAdmin.PATCH("/users/:id/toggle-status", authHandler.ToggleUserStatus)
				authAdmin.POST("/users/:id/reset-password", authHandler.ResetUserPassword)
			}
		}

		// Products routes (PHASE 3)
		products := v1.Group("/products")
		products.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		// Rate limit: QR search adalah endpoint kritikal (<500ms target).
		// Batasi 60 req/menit/IP supaya tidak di-spam.
		products.Use(middleware.RateLimiter(cfg.RedisClient, 60, time.Minute))
		{
			products.GET("", productHandler.ListProducts)
			products.GET("/code/:code", productHandler.GetProductByCode) // Kritikal untuk scan (QR = product_code)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/:id/stock-history", stockOpnameHandler.GetProductTrend)

			// Admin only
			productsAdmin := products.Group("")
			productsAdmin.Use(middleware.RequireRole("admin"))
			{
				productsAdmin.POST("", productHandler.CreateProduct)
				productsAdmin.PUT("/:id", productHandler.UpdateProduct)
				productsAdmin.DELETE("/:id", productHandler.DeleteProduct)
				productsAdmin.POST("/transfer", productHandler.TransferProducts)
			}
		}

		// Categories routes
		categories := v1.Group("/categories")
		categories.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			categories.GET("", categoryHandler.ListCategories)

			// Admin only
			categoriesAdmin := categories.Group("")
			categoriesAdmin.Use(middleware.RequireRole("admin"))
			{
				categoriesAdmin.POST("", categoryHandler.CreateCategory)
			}
		}

		// Suppliers routes
		suppliers := v1.Group("/suppliers")
		suppliers.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			suppliers.GET("", supplierHandler.ListSuppliers)

			// Admin only
			suppliersAdmin := suppliers.Group("")
			suppliersAdmin.Use(middleware.RequireRole("admin"))
			{
				suppliersAdmin.POST("", supplierHandler.CreateSupplier)
			}
		}

		// Warehouses routes
		warehouses := v1.Group("/warehouses")
		warehouses.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			warehouses.GET("", warehouseHandler.ListWarehouses)
			warehouses.GET("/:id/locations", warehouseHandler.GetLocations)
			warehouses.GET("/:id/layout", warehouseHandler.GetLayout(cfg.DB.DB))

			// Admin only
			warehousesAdmin := warehouses.Group("")
			warehousesAdmin.Use(middleware.RequireRole("admin"))
			{
				warehousesAdmin.POST("", warehouseHandler.CreateWarehouse)
				warehousesAdmin.PUT("/:id", warehouseHandler.UpdateWarehouse)
				warehousesAdmin.DELETE("/:id", warehouseHandler.DeleteWarehouse)
				warehousesAdmin.POST("/:id/locations", warehouseHandler.CreateLocation)
				warehousesAdmin.PUT("/:id/layout", warehouseHandler.UpdateLayout(cfg.DB.DB))
			}
		}

		// Locations routes
		locations := v1.Group("/locations")
		locations.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			locations.GET("/:id/boxes", warehouseHandler.GetBoxes)

			// Admin only
			locationsAdmin := locations.Group("")
			locationsAdmin.Use(middleware.RequireRole("admin"))
			{
				locationsAdmin.POST("/:id/boxes", warehouseHandler.CreateBox)
				locationsAdmin.PUT("/:id", warehouseHandler.UpdateLocation)
				locationsAdmin.DELETE("/:id", warehouseHandler.DeleteLocation)
			}
		}

		// Boxes routes (for direct update/delete)
		boxes := v1.Group("/boxes")
		boxes.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			// Admin only
			boxesAdmin := boxes.Group("")
			boxesAdmin.Use(middleware.RequireRole("admin"))
			{
				boxesAdmin.PUT("/:id", warehouseHandler.UpdateBox)
				boxesAdmin.DELETE("/:id", warehouseHandler.DeleteBox)
			}
		}

		// Stock opname routes (PHASE 4)
		stockOpname := v1.Group("/stock-opname")
		stockOpname.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			stockOpname.GET("/sessions", stockOpnameHandler.ListSessions)
			stockOpname.POST("/sessions", stockOpnameHandler.StartSession)
			stockOpname.GET("/sessions/:id", stockOpnameHandler.GetSession)
			stockOpname.GET("/sessions/:id/details", stockOpnameHandler.GetSessionDetails)
			stockOpname.GET("/sessions/:id/summary", stockOpnameHandler.GetSessionSummary)
			stockOpname.POST("/sessions/:id/scan", stockOpnameHandler.SubmitScan)
			stockOpname.POST("/sessions/:id/complete", stockOpnameHandler.CompleteSession)

			// Supervisor only
			stockOpnameSupervisor := stockOpname.Group("")
			stockOpnameSupervisor.Use(middleware.RequireRole("supervisor", "admin"))
			{
				stockOpnameSupervisor.POST("/sessions/:id/approve", stockOpnameHandler.ApproveSession)
				stockOpnameSupervisor.POST("/sessions/:id/reject", stockOpnameHandler.RejectSession)
			}

			// Admin only: reopen
			stockOpnameAdmin := stockOpname.Group("")
			stockOpnameAdmin.Use(middleware.RequireRole("admin"))
			{
				stockOpnameAdmin.POST("/sessions/:id/reopen", stockOpnameHandler.ReopenSession)
			}
		}

		// Multi-warehouse reconciliation (admin & supervisor)
		reconciliation := v1.Group("/reconciliation")
		reconciliation.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			reconciliationSupervisor := reconciliation.Group("")
			reconciliationSupervisor.Use(middleware.RequireRole("supervisor", "admin"))
			{
				reconciliationSupervisor.POST("", reconciliationHandler.ReconcileWarehouses)
			}
		}

		// Stock snapshot comparison
		snapshots := v1.Group("/snapshots")
		snapshots.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			snapshots.GET("/compare", snapshotHandler.GetSnapshotComparison)
		}

		// Stock Ledger (kartu stok) - supervisor & admin
		stockLedger := v1.Group("/stock-ledger")
		stockLedger.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		stockLedger.Use(middleware.RequireRole("supervisor", "admin"))
		{
			stockLedger.GET("/product/:product_id/:warehouse_id", stockLedgerHandler.GetProductLedger)
			stockLedger.GET("/warehouse/:warehouse_id", stockLedgerHandler.GetWarehouseLedger)
			stockLedger.GET("/range", stockLedgerHandler.GetLedgerByDateRange)
		}

		// Manual stock adjustment (admin & supervisor)
		stocks := v1.Group("/stocks")
		stocks.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		// Rate limit: stock adjustment adalah operasi kritis, 20x/menit/IP.
		stocks.Use(middleware.RateLimiter(cfg.RedisClient, 20, time.Minute))
		{
			stocksAdmin := stocks.Group("")
			stocksAdmin.Use(middleware.RequireRole("supervisor", "admin"))
			{
				stocksAdmin.POST("/adjust", stockAdjustmentHandler.AdjustStock)
			}
		}

		// Inventory movement (Stock IN/OUT/transfer antar gudang) — supervisor & admin.
		// Operasi berat (ubah stok + tulis ledger), batasi 30 req/menit/IP.
		inventory := v1.Group("/inventory")
		inventory.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		inventory.Use(middleware.RateLimiter(cfg.RedisClient, 30, time.Minute))
		{
			inventory.GET("", inventoryHandler.ListMovements)
			inventory.GET("/:id", inventoryHandler.GetMovement)

			inventoryWrite := inventory.Group("")
			inventoryWrite.Use(middleware.RequireRole("supervisor", "admin"))
			{
				inventoryWrite.POST("/in", inventoryHandler.PostIn)
				inventoryWrite.POST("/out", inventoryHandler.PostOut)
				inventoryWrite.POST("/transfer", inventoryHandler.PostTransfer)
				inventoryWrite.POST("/grid", inventoryHandler.PostGrid)
			}
			// Read endpoints grid + summary (supervisor & admin).
			inventory.GET("/grid", inventoryHandler.GetGrid)
			inventory.GET("/summary", inventoryHandler.GetSummary)
		}

		// Backup & restore (admin only)
		backup := v1.Group("/backup")
		backup.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		backup.Use(middleware.RequireRole("admin"))
		{
			backup.GET("/status", backupHandler.GetBackupStatus)
			backup.GET("/download", backupHandler.TriggerBackup)
		}

		// WebSocket for real-time progress (PHASE 4).
		// Token dikirim lewat sub-protokol handshake, bukan query string, dan
		// keanggotaan sesi diperiksa sebelum upgrade.
		v1.GET("/ws/stock-opname/sessions/:id", websocket.HandleStockOpnameWS(&websocket.WSConfig{
			JWTSecret:         cfg.Config.JWT.AccessSecret,
			AllowedOrigins:    strings.Split(cfg.CORSAllowedOrigins, ","),
			CanAccessSession:  canAccessSession(stockOpnameUsecase),
		}))

		// Dashboard routes (PHASE 4)
		dashboard := v1.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			dashboard.GET("/summary", dashboardHandler.GetSummary)
		}

		// Reports routes (PHASE 5) - Supervisor/Admin only
		reports := v1.Group("/reports")
		reports.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		reports.Use(middleware.RequireRole("supervisor", "admin"))
		{
			reports.GET("/stock-opname", reportHandler.GetStockOpnameReport)
			reports.GET("/stock-opname/export", reportHandler.ExportStockOpnameReport)
		}

		// Activity logs routes (PHASE 5) - Supervisor/Admin only
		activityLogs := v1.Group("/activity-logs")
		activityLogs.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		activityLogs.Use(middleware.RequireRole("supervisor", "admin"))
		{
			activityLogs.GET("", reportHandler.GetActivityLogs)
		}

		// Notifications routes (PHASE 5) - All authenticated users
		notifications := v1.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		{
			notifications.GET("", reportHandler.GetNotifications)
			notifications.PATCH("/:id/read", reportHandler.MarkNotificationAsRead)
		}

		// Export routes - Supervisor/Admin only (data seluruh gudang)
		export := v1.Group("/export")
		export.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		export.Use(middleware.RequireRole("supervisor", "admin"))
		{
			export.GET("/stock-opname", exportHandler.ExportStockOpnameExcel)
			export.GET("/products", exportHandler.ExportProductsExcel)
		}

		// Import routes (produk via template Excel) - Admin only
		imports := v1.Group("/import")
		imports.Use(middleware.AuthMiddleware(cfg.Config.JWT.AccessSecret))
		imports.Use(middleware.RequireRole("admin"))
		// Rate limit: import adalah operasi berat, 5x/5menit/IP.
		imports.Use(middleware.RateLimiter(cfg.RedisClient, 5, 5*time.Minute))
		{
			imports.GET("/products/template", importHandler.DownloadProductTemplate)
			imports.POST("/products", importHandler.ImportProducts)
		}

		// Placeholder untuk memastikan /api/v1 dapat diakses
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}
}
// ===== END FILE =====
