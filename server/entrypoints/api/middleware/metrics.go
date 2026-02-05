package middleware

import (
	"net/http"
	"time"

	"github.com/kamilrybacki/claudeception/server/services/metrics"
)

// Metrics middleware records HTTP request metrics
type Metrics struct {
	service metrics.Service
}

// NewMetrics creates a new metrics middleware
func NewMetrics(service metrics.Service) *Metrics {
	return &Metrics{service: service}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Middleware returns an HTTP handler that records request metrics
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		userID := GetUserID(r.Context())

		m.service.RecordAPIRequest(
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
			userID,
		)
	})
}
