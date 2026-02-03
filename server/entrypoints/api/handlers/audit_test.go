package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/claudeception/server/services/audit"
)

type mockAuditService struct {
	entries []domain.AuditEntry
}

func (m *mockAuditService) List(ctx context.Context, params audit.ListParams) ([]domain.AuditEntry, int, error) {
	return m.entries, len(m.entries), nil
}

func (m *mockAuditService) GetEntityHistory(ctx context.Context, entityType domain.AuditEntityType, entityID string) ([]domain.AuditEntry, error) {
	var result []domain.AuditEntry
	for _, e := range m.entries {
		if e.EntityType == entityType && e.EntityID == entityID {
			result = append(result, e)
		}
	}
	return result, nil
}

func TestAuditHandler_List(t *testing.T) {
	actorID := "user-1"
	svc := &mockAuditService{
		entries: []domain.AuditEntry{
			{
				ID:         "entry-1",
				EntityType: domain.AuditEntityRule,
				EntityID:   "rule-1",
				Action:     domain.AuditActionCreated,
				ActorID:    &actorID,
				CreatedAt:  time.Now(),
			},
		},
	}
	h := handlers.NewAuditHandler(svc)

	req := httptest.NewRequest("GET", "/audit", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handlers.AuditListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp.Entries))
	}
}

func TestAuditHandler_GetEntityHistory(t *testing.T) {
	actorID := "user-1"
	svc := &mockAuditService{
		entries: []domain.AuditEntry{
			{
				ID:         "entry-1",
				EntityType: domain.AuditEntityRule,
				EntityID:   "rule-1",
				Action:     domain.AuditActionCreated,
				ActorID:    &actorID,
				CreatedAt:  time.Now().Add(-time.Hour),
			},
			{
				ID:         "entry-2",
				EntityType: domain.AuditEntityRule,
				EntityID:   "rule-1",
				Action:     domain.AuditActionUpdated,
				ActorID:    &actorID,
				CreatedAt:  time.Now(),
			},
		},
	}
	h := handlers.NewAuditHandler(svc)

	// Create a chi router to handle URL params
	r := chi.NewRouter()
	r.Get("/audit/entity/{entityType}/{entityId}", h.GetEntityHistory)

	req := httptest.NewRequest("GET", "/audit/entity/rule/rule-1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []handlers.AuditEntryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp))
	}
}
