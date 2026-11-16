// ===== FILE: backend/internal/delivery/http/handler/report_filter.go =====
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

// parseReportFilter membaca filter laporan dari query string. Dipakai bersama
// oleh JSON dan CSV/Excel supaya keduanya menyaring dengan aturan yang sama.
// Nilai yang tidak valid mengembalikan 400, bukan diabaikan diam-diam.
func parseReportFilter(c *gin.Context) (usecase.StockOpnameReportFilter, error) {
	var f usecase.StockOpnameReportFilter
	f.Brand = c.Query("brand")

	var err error
	if f.SessionID, err = parseOptionalUUID(c, "session_id"); err != nil {
		return f, err
	}
	if f.WarehouseID, err = parseOptionalUUID(c, "warehouse_id"); err != nil {
		return f, err
	}
	if f.SupplierID, err = parseOptionalUUID(c, "supplier_id"); err != nil {
		return f, err
	}
	if f.UserID, err = parseOptionalUUID(c, "user_id"); err != nil {
		return f, err
	}
	if f.DateFrom, err = parseOptionalDate(c, "date_from"); err != nil {
		return f, err
	}
	if f.DateTo, err = parseOptionalDate(c, "date_to"); err != nil {
		return f, err
	}
	if f.DateFrom != nil && f.DateTo != nil && f.DateTo.Before(*f.DateFrom) {
		return f, errors.New("date_to tidak boleh sebelum date_from")
	}
	if f.MinDiff, err = parseOptionalInt(c, "min_diff"); err != nil {
		return f, err
	}
	if f.MaxDiff, err = parseOptionalInt(c, "max_diff"); err != nil {
		return f, err
	}
	if f.MinDiff != nil && f.MaxDiff != nil && *f.MaxDiff < *f.MinDiff {
		return f, errors.New("max_diff tidak boleh lebih kecil dari min_diff")
	}
	return f, nil
}

func parseOptionalUUID(c *gin.Context, key string) (*uuid.UUID, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s bukan UUID yang valid", key)
	}
	return &id, nil
}

func parseOptionalDate(c *gin.Context, key string) (*time.Time, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, fmt.Errorf("%s harus format YYYY-MM-DD", key)
	}
	return &t, nil
}

func parseOptionalInt(c *gin.Context, key string) (*int, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("%s harus bilangan bulat", key)
	}
	return &v, nil
}

// respondFilterError memetakan kesalahan parse filter ke 400.
func respondFilterError(c *gin.Context, err error) {
	utils.RespondError(c, http.StatusBadRequest, "INVALID_FILTER", err.Error())
}

// ===== END FILE =====
