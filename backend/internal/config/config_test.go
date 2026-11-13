package config

import (
	"strings"
	"testing"
	"time"
)

// Regresi 2.10: secret kosong, terlalu pendek, atau nilai contoh harus ditolak.
func TestValidateSecret(t *testing.T) {
	strong := strings.Repeat("a1b2c3d4", 6) // 48 karakter, tanpa kata terlarang

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"kosong", "", true},
		{"terlalu pendek", "abc123", true},
		{"nilai contoh", "change-this-to-a-random-secret-key-in-production-xx", true},
		{"kuat", strong, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecret("JWT_ACCESS_SECRET", tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateSecret() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Regresi CQ-01: nama variabel dari .env.example harus benar-benar terbaca.
func TestLoadReadsEnvExampleKeys(t *testing.T) {
	t.Setenv("JWT_ACCESS_SECRET", strings.Repeat("a1b2c3d4", 6))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("e5f6g7h8", 6))
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("ENV", "development")
	t.Setenv("JWT_ACCESS_EXPIRY", "20")
	t.Setenv("JWT_REFRESH_EXPIRY", "10080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() err = %v", err)
	}
	if cfg.App.Port != "9000" {
		t.Errorf("Port = %q, want 9000", cfg.App.Port)
	}
	if cfg.App.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.App.Env)
	}
	if cfg.JWT.AccessTTL != 20*time.Minute {
		t.Errorf("AccessTTL = %v, want 20m", cfg.JWT.AccessTTL)
	}
	if cfg.JWT.RefreshTTL != 10080*time.Minute {
		t.Errorf("RefreshTTL = %v, want 10080m (7 hari)", cfg.JWT.RefreshTTL)
	}
}

// Nilai numerik yang tidak valid harus gagal start, bukan jatuh ke default.
func TestLoadRejectsNonNumericExpiry(t *testing.T) {
	t.Setenv("JWT_ACCESS_SECRET", strings.Repeat("a1b2c3d4", 6))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("e5f6g7h8", 6))
	t.Setenv("JWT_ACCESS_EXPIRY", "lima belas")

	if _, err := Load(); err == nil {
		t.Fatal("Load() err = nil, want error untuk JWT_ACCESS_EXPIRY non-angka")
	}
}

// ===== END FILE =====
