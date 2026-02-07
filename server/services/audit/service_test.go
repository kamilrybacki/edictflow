package audit

import (
	"context"
	"testing"
	"time"

	"github.com/kamilrybacki/edictflow/server/adapters/postgres"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type mockAuditDB struct {
	entries []domain.AuditEntry
}

func (m *mockAuditDB) Create(ctx context.Context, entry domain.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditDB) List(ctx context.Context, params postgres.AuditListParams) ([]domain.AuditEntry, int, error) {
	var result []domain.AuditEntry
	for _, e := range m.entries {
		if params.EntityType != nil && e.EntityType != *params.EntityType {
			continue
		}
		if params.EntityID != nil && e.EntityID != *params.EntityID {
			continue
		}
		if params.ActorID != nil && (e.ActorID == nil || *e.ActorID != *params.ActorID) {
			continue
		}
		if params.Action != nil && e.Action != *params.Action {
			continue
		}
		result = append(result, e)
	}

	// Apply pagination
	total := len(result)
	if params.Offset > 0 && params.Offset < len(result) {
		result = result[params.Offset:]
	}
	if params.Limit > 0 && params.Limit < len(result) {
		result = result[:params.Limit]
	}

	return result, total, nil
}

func (m *mockAuditDB) GetEntityHistory(ctx context.Context, entityType domain.AuditEntityType, entityID string) ([]domain.AuditEntry, error) {
	var result []domain.AuditEntry
	for _, e := range m.entries {
		if e.EntityType == entityType && e.EntityID == entityID {
			result = append(result, e)
		}
	}
	return result, nil
}

func newTestService() (*Service, *mockAuditDB) {
	db := &mockAuditDB{entries: []domain.AuditEntry{}}
	return NewService(db), db
}

func TestService_LogCreate(t *testing.T) {
	svc, db := newTestService()
	actorID := "user-1"

	err := svc.LogCreate(context.Background(), domain.AuditEntityRule, "rule-1", &actorID, nil)
	if err != nil {
		t.Fatalf("LogCreate() error = %v", err)
	}

	if len(db.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(db.entries))
	}

	entry := db.entries[0]
	if entry.EntityType != domain.AuditEntityRule {
		t.Errorf("Expected entity type 'rule', got '%s'", entry.EntityType)
	}
	if entry.EntityID != "rule-1" {
		t.Errorf("Expected entity ID 'rule-1', got '%s'", entry.EntityID)
	}
	if entry.Action != domain.AuditActionCreated {
		t.Errorf("Expected action 'created', got '%s'", entry.Action)
	}
	if entry.ActorID == nil || *entry.ActorID != actorID {
		t.Errorf("Expected actor ID '%s', got '%v'", actorID, entry.ActorID)
	}
}

func TestService_LogUpdate(t *testing.T) {
	svc, db := newTestService()
	actorID := "user-1"
	changes := map[string]*domain.ChangeValue{
		"name": {Old: "Old Name", New: "New Name"},
	}

	err := svc.LogUpdate(context.Background(), domain.AuditEntityRule, "rule-1", &actorID, changes, nil)
	if err != nil {
		t.Fatalf("LogUpdate() error = %v", err)
	}

	if len(db.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(db.entries))
	}

	entry := db.entries[0]
	if entry.Action != domain.AuditActionUpdated {
		t.Errorf("Expected action 'updated', got '%s'", entry.Action)
	}
	if len(entry.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(entry.Changes))
	}
	if entry.Changes["name"].Old != "Old Name" || entry.Changes["name"].New != "New Name" {
		t.Errorf("Changes not recorded correctly")
	}
}

func TestService_LogDelete(t *testing.T) {
	svc, db := newTestService()
	actorID := "user-1"

	err := svc.LogDelete(context.Background(), domain.AuditEntityRule, "rule-1", &actorID, nil)
	if err != nil {
		t.Fatalf("LogDelete() error = %v", err)
	}

	if len(db.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(db.entries))
	}

	entry := db.entries[0]
	if entry.Action != domain.AuditActionDeleted {
		t.Errorf("Expected action 'deleted', got '%s'", entry.Action)
	}
}

func TestService_LogApprovalAction(t *testing.T) {
	svc, db := newTestService()
	actorID := "approver-1"
	metadata := map[string]interface{}{"comment": "LGTM"}

	err := svc.LogApprovalAction(context.Background(), "rule-1", domain.AuditActionApproved, &actorID, metadata)
	if err != nil {
		t.Fatalf("LogApprovalAction() error = %v", err)
	}

	if len(db.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(db.entries))
	}

	entry := db.entries[0]
	if entry.EntityType != domain.AuditEntityRule {
		t.Errorf("Expected entity type 'rule', got '%s'", entry.EntityType)
	}
	if entry.Action != domain.AuditActionApproved {
		t.Errorf("Expected action 'approved', got '%s'", entry.Action)
	}
	if entry.Metadata["comment"] != "LGTM" {
		t.Errorf("Expected metadata comment 'LGTM', got '%v'", entry.Metadata["comment"])
	}
}

func TestService_List(t *testing.T) {
	svc, db := newTestService()

	// Add some entries
	actorID := "user-1"
	db.entries = append(db.entries, domain.AuditEntry{
		ID:         "entry-1",
		EntityType: domain.AuditEntityRule,
		EntityID:   "rule-1",
		Action:     domain.AuditActionCreated,
		ActorID:    &actorID,
		CreatedAt:  time.Now(),
	})
	db.entries = append(db.entries, domain.AuditEntry{
		ID:         "entry-2",
		EntityType: domain.AuditEntityUser,
		EntityID:   "user-2",
		Action:     domain.AuditActionCreated,
		ActorID:    &actorID,
		CreatedAt:  time.Now(),
	})

	// List all
	entries, total, err := svc.List(context.Background(), ListParams{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Filter by entity type
	entityType := domain.AuditEntityRule
	entries, filteredTotal, err := svc.List(context.Background(), ListParams{EntityType: &entityType})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if filteredTotal != 1 {
		t.Errorf("Expected total 1, got %d", filteredTotal)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestService_GetEntityHistory(t *testing.T) {
	svc, db := newTestService()
	actorID := "user-1"

	db.entries = append(db.entries, domain.AuditEntry{
		ID:         "entry-1",
		EntityType: domain.AuditEntityRule,
		EntityID:   "rule-1",
		Action:     domain.AuditActionCreated,
		ActorID:    &actorID,
		CreatedAt:  time.Now().Add(-time.Hour),
	})
	db.entries = append(db.entries, domain.AuditEntry{
		ID:         "entry-2",
		EntityType: domain.AuditEntityRule,
		EntityID:   "rule-1",
		Action:     domain.AuditActionUpdated,
		ActorID:    &actorID,
		CreatedAt:  time.Now(),
	})
	db.entries = append(db.entries, domain.AuditEntry{
		ID:         "entry-3",
		EntityType: domain.AuditEntityRule,
		EntityID:   "rule-2",
		Action:     domain.AuditActionCreated,
		ActorID:    &actorID,
		CreatedAt:  time.Now(),
	})

	entries, err := svc.GetEntityHistory(context.Background(), domain.AuditEntityRule, "rule-1")
	if err != nil {
		t.Fatalf("GetEntityHistory() error = %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries for rule-1, got %d", len(entries))
	}
}
