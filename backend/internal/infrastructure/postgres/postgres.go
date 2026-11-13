// ===== FILE: backend/internal/infrastructure/postgres/postgres.go =====
package postgres

import (
	"fmt"
	"time"

	"github.com/nyouaku/SOPQ/backend/internal/config"
	"github.com/nyouaku/SOPQ/backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

// NewPostgresDB membuat koneksi ke PostgreSQL dengan GORM
func NewPostgresDB(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := cfg.DSN()

	// Konfigurasi GORM logger sesuai environment
	logLevel := gormlogger.Silent
	if cfg.Host == "localhost" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test koneksi
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("PostgreSQL connected successfully")

	return &Database{DB: db}, nil
}

// Close menutup koneksi database
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// RunMigrations ada di migrate.go.
// ===== END FILE =====
