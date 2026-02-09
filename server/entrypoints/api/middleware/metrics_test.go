package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockMetricsService struct {
	recordedMethod   string
	recordedPath     string
	recordedStatus   int
	recordedDuration time.Duration
	recordedUserID   string
}

func (m *mockMetricsService) RecordAPIRequest(method, path string, statusCode int, duration time.Duration, userID string) {
	m.recordedMethod = method
	m.recordedPath = path
	m.recordedStatus = statusCode
	m.recordedDuration = duration
	m.recordedUserID = userID
}

func (m *mockMetricsService) RecordAPIError(method, path string, errorType string, userID string)      {}
func (m *mockMetricsService) RecordRedisEvent(eventType string, channel string, success bool)         {}
func (m *mockMetricsService) RecordRedisPublish(channel, messageType string, success bool, latencyMs int64) {
}
func (m *mockMetricsService) RecordRedisSubscription(channel, action string)                          {}
func (m *mockMetricsService) RecordHubStats(agents, teams, subscriptions int)                         {}
func (m *mockMetricsService) RecordAgentConnection(agentID, teamID, action string)                    {}
func (m *mockMetricsService) RecordWebSocketMessage(direction, messageType, agentID string, sizeBytes int) {
}
func (m *mockMetricsService) RecordBroadcast(teamID, eventType string, recipientCount int) {}
func (m *mockMetricsService) RecordDBQuery(operation string, table string, duration time.Duration, success bool) {
}
func (m *mockMetricsService) RecordDBPoolStats(totalConns, acquiredConns, idleConns, maxConns int32) {}
func (m *mockMetricsService) RecordHealthCheck(component, status string, latencyMs int64)            {}
func (m *mockMetricsService) RecordWorkerHeartbeat(workerID string, agentCount, teamCount int)       {}
func (m *mockMetricsService) Flush() error                                                            { return nil }
func (m *mockMetricsService) Close() error                                                            { return nil }

func TestMetricsMiddleware(t *testing.T) {
	mock := &mockMetricsService{}
	metricsMiddleware := NewMetrics(mock)

	handler := metricsMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if mock.recordedMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", mock.recordedMethod)
	}

	if mock.recordedPath != "/api/v1/rules" {
		t.Errorf("expected path /api/v1/rules, got %s", mock.recordedPath)
	}

	if mock.recordedStatus != http.StatusOK {
		t.Errorf("expected status 200, got %d", mock.recordedStatus)
	}

	if mock.recordedDuration <= 0 {
		t.Error("expected duration > 0")
	}
}

func TestMetricsMiddlewareWithError(t *testing.T) {
	mock := &mockMetricsService{}
	metricsMiddleware := NewMetrics(mock)

	handler := metricsMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rules", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if mock.recordedMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", mock.recordedMethod)
	}

	if mock.recordedStatus != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", mock.recordedStatus)
	}
}

func TestMetricsMiddlewareWithUserID(t *testing.T) {
	mock := &mockMetricsService{}
	metricsMiddleware := NewMetrics(mock)

	handler := metricsMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-123")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if mock.recordedUserID != "user-123" {
		t.Errorf("expected userID user-123, got %s", mock.recordedUserID)
	}
}

func TestMetricsMiddlewareDefaultStatus(t *testing.T) {
	mock := &mockMetricsService{}
	metricsMiddleware := NewMetrics(mock)

	handler := metricsMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write body without explicitly calling WriteHeader
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if mock.recordedStatus != http.StatusOK {
		t.Errorf("expected default status 200, got %d", mock.recordedStatus)
	}
}

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	// Write header once
	rw.WriteHeader(http.StatusNotFound)

	// Try to write again (should be ignored)
	rw.WriteHeader(http.StatusOK)

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404 to be preserved, got %d", rw.statusCode)
	}
}
