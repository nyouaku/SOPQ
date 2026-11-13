// ===== FILE: backend/internal/middleware/metrics.go =====
package middleware

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrik minimal dalam format Prometheus text (tanpa dependency baru).
// Label memakai route template (c.FullPath()), bukan path mentah, supaya
// kardinalitas tetap terbatas.

type metricKey struct {
	method string
	route  string
	status int
}

type metricVal struct {
	count      uint64
	sumSeconds float64
}

var (
	metricsMu sync.Mutex
	metrics   = map[metricKey]*metricVal{}
)

func recordRequest(method, route string, status int, d time.Duration) {
	if route == "" {
		route = "unmatched"
	}
	k := metricKey{method: method, route: route, status: status}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	v, ok := metrics[k]
	if !ok {
		v = &metricVal{}
		metrics[k] = v
	}
	v.count++
	v.sumSeconds += d.Seconds()
}

// MetricsHandler menulis metrik dalam format exposition Prometheus.
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsMu.Lock()
		keys := make([]metricKey, 0, len(metrics))
		for k := range metrics {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, b := keys[i], keys[j]
			if a.route != b.route {
				return a.route < b.route
			}
			if a.method != b.method {
				return a.method < b.method
			}
			return a.status < b.status
		})

		var sb strings.Builder
		sb.WriteString("# HELP http_requests_total Jumlah request HTTP.\n")
		sb.WriteString("# TYPE http_requests_total counter\n")
		for _, k := range keys {
			fmt.Fprintf(&sb, "http_requests_total{method=%q,route=%q,status=\"%d\"} %d\n",
				k.method, k.route, k.status, metrics[k].count)
		}
		sb.WriteString("# HELP http_request_duration_seconds_sum Total latensi request (detik).\n")
		sb.WriteString("# TYPE http_request_duration_seconds_sum counter\n")
		for _, k := range keys {
			fmt.Fprintf(&sb, "http_request_duration_seconds_sum{method=%q,route=%q,status=\"%d\"} %f\n",
				k.method, k.route, k.status, metrics[k].sumSeconds)
		}
		metricsMu.Unlock()

		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(sb.String()))
	}
}

// ===== END FILE =====
