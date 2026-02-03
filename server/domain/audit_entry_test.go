package domain

import (
	"testing"
)

func TestAuditEntry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		entry   AuditEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: AuditEntry{
				EntityType: AuditEntityRule,
				EntityID:   "rule-123",
				Action:     AuditActionCreated,
				ActorID:    strPtr("user-123"),
			},
			wantErr: false,
		},
		{
			name: "empty entity type",
			entry: AuditEntry{
				EntityType: "",
				EntityID:   "rule-123",
				Action:     AuditActionCreated,
			},
			wantErr: true,
		},
		{
			name: "empty entity id",
			entry: AuditEntry{
				EntityType: AuditEntityRule,
				EntityID:   "",
				Action:     AuditActionCreated,
			},
			wantErr: true,
		},
		{
			name: "empty action",
			entry: AuditEntry{
				EntityType: AuditEntityRule,
				EntityID:   "rule-123",
				Action:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewAuditEntry(t *testing.T) {
	actorID := "user-123"
	entry := NewAuditEntry(AuditEntityRule, "rule-456", AuditActionCreated, &actorID)

	if entry.ID == "" {
		t.Error("Expected entry to have an ID")
	}
	if entry.EntityType != AuditEntityRule {
		t.Errorf("Expected EntityType 'rule', got %s", entry.EntityType)
	}
	if entry.EntityID != "rule-456" {
		t.Errorf("Expected EntityID 'rule-456', got %s", entry.EntityID)
	}
	if entry.Changes == nil {
		t.Error("Expected Changes map to be initialized")
	}
	if entry.Metadata == nil {
		t.Error("Expected Metadata map to be initialized")
	}
}

func TestAuditEntry_AddChange(t *testing.T) {
	entry := NewAuditEntry(AuditEntityRule, "rule-123", AuditActionUpdated, nil)
	entry.AddChange("name", "Old Name", "New Name")

	if len(entry.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(entry.Changes))
	}
	change, ok := entry.Changes["name"]
	if !ok {
		t.Error("Expected 'name' change to exist")
	}
	if change.Old != "Old Name" || change.New != "New Name" {
		t.Error("Change values don't match")
	}
}

func TestAuditEntry_AddMetadata(t *testing.T) {
	entry := NewAuditEntry(AuditEntityRule, "rule-123", AuditActionApproved, nil)
	entry.AddMetadata("approval_count", 3)

	if len(entry.Metadata) != 1 {
		t.Errorf("Expected 1 metadata entry, got %d", len(entry.Metadata))
	}
	if entry.Metadata["approval_count"] != 3 {
		t.Error("Metadata value doesn't match")
	}
}

func strPtr(s string) *string {
	return &s
}
