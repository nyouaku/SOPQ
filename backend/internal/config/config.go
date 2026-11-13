// ===== FILE: backend/internal/config/config.go =====
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Misc     MiscConfig
}

type AppConfig struct {
	Env  string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type JWTConfig struct {
	AccessSecret     string
	RefreshSecret    string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
}

type MiscConfig struct {
	DiffWarningThresholdPercent int
	UploadDir                   string
	CORSAllowedOrigins          string
}

// Load membaca .env file dan mengembalikan Config.
// Nama variabel mengikuti .env.example. Nilai yang tidak bisa di-parse menghasilkan
// error, bukan diam-diam jatuh ke default (CQ-01).
func Load() (*Config, error) {
	// Load .env file (optional, tidak error jika tidak ada karena bisa pakai env vars langsung)
	_ = godotenv.Load()

	var errs []string
	intEnv := func(key string, def int) int {
		v, err := envInt(key, def)
		if err != nil {
			errs = append(errs, err.Error())
		}
		return v
	}

	cfg := &Config{
		App: AppConfig{
			// ENV dipakai .env.example; APP_ENV tetap diterima sebagai alias.
			Env:  envString("APP_ENV", envString("ENV", "development")),
			Port: envString("SERVER_PORT", envString("APP_PORT", "8080")),
		},
		Database: DatabaseConfig{
			Host:     envString("DB_HOST", "localhost"),
			Port:     envString("DB_PORT", "5432"),
			User:     envString("DB_USER", "postgres"),
			Password: envString("DB_PASSWORD", ""),
			DBName:   envString("DB_NAME", "stockopname_db"),
			SSLMode:  envString("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     envString("REDIS_HOST", "localhost"),
			Port:     envString("REDIS_PORT", "6379"),
			Password: envString("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			AccessSecret:  envString("JWT_ACCESS_SECRET", ""),
			RefreshSecret: envString("JWT_REFRESH_SECRET", ""),
			// .env.example: JWT_ACCESS_EXPIRY dan JWT_REFRESH_EXPIRY keduanya dalam menit.
			// JWT_*_TTL_* (menit / hari) tetap diterima sebagai alias lama.
			AccessTTL:  time.Duration(intEnv("JWT_ACCESS_EXPIRY", intEnv("JWT_ACCESS_TTL_MINUTES", 15))) * time.Minute,
			RefreshTTL: time.Duration(intEnv("JWT_REFRESH_EXPIRY", intEnv("JWT_REFRESH_TTL_DAYS", 7)*1440)) * time.Minute,
		},
		Misc: MiscConfig{
			DiffWarningThresholdPercent: intEnv("DIFF_WARNING_THRESHOLD_PERCENT", 10),
			UploadDir:                   envString("UPLOAD_DIR", "./uploads"),
			CORSAllowedOrigins:          envString("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		},
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("konfigurasi tidak valid: %s", strings.Join(errs, "; "))
	}

	// Validasi config wajib.
	// Secret default / terlalu pendek berarti siapa pun yang tahu nilai contoh
	// di .env.example bisa menempa token. Gagal di startup, bukan di produksi.
	if err := validateSecret("JWT_ACCESS_SECRET", cfg.JWT.AccessSecret); err != nil {
		return nil, err
	}
	if err := validateSecret("JWT_REFRESH_SECRET", cfg.JWT.RefreshSecret); err != nil {
		return nil, err
	}
	if cfg.JWT.AccessSecret == cfg.JWT.RefreshSecret {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET dan JWT_REFRESH_SECRET harus berbeda")
	}

	if cfg.App.Env == "production" && cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD wajib diisi saat APP_ENV=production")
	}

	return cfg, nil
}

// minSecretLength 32 byte = 256 bit, batas bawah yang wajar untuk HS256.
const minSecretLength = 32

func validateSecret(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s wajib diisi", name)
	}
	if len(value) < minSecretLength {
		return fmt.Errorf("%s terlalu pendek: minimal %d karakter (sekarang %d)", name, minSecretLength, len(value))
	}
	if strings.Contains(strings.ToLower(value), "change-this") ||
		strings.Contains(strings.ToLower(value), "changeme") ||
		strings.Contains(strings.ToLower(value), "secret-key") {
		return fmt.Errorf("%s masih memakai nilai contoh; ganti dengan nilai acak", name)
	}
	return nil
}

func envString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// envInt: variabel kosong → default; variabel terisi tapi bukan angka → error.
func envInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s harus angka, dapat %q", key, value)
	}
	return intVal, nil
}

// DSN mengembalikan Data Source Name untuk PostgreSQL
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// RedisAddr mengembalikan alamat Redis dalam format host:port
func (c *RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
// ===== END FILE =====
