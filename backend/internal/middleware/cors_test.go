package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Regresi 2.4: wildcard tidak boleh disertai Allow-Credentials, dan origin
// yang tidak terdaftar tidak boleh mendapat Allow-Origin.
func TestCORS_OriginAllowList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		allowed     string
		origin      string
		wantAllow   string
		wantCredent string
	}{
		{"terdaftar", "http://localhost:3000,http://localhost:80", "http://localhost:3000", "http://localhost:3000", "true"},
		{"tidak terdaftar", "http://localhost:3000", "https://evil.example", "", ""},
		{"wildcard tanpa credentials", "*", "https://anything.example", "https://anything.example", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(CORS(tt.allowed))
			r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if got := w.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllow {
				t.Errorf("Allow-Origin = %q, want %q", got, tt.wantAllow)
			}
			if got := w.Header().Get("Access-Control-Allow-Credentials"); got != tt.wantCredent {
				t.Errorf("Allow-Credentials = %q, want %q", got, tt.wantCredent)
			}
		})
	}
}

// ===== END FILE =====
