package domain

import (
	"testing"
)

func TestRole_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    RoleEntity
		wantErr bool
	}{
		{
			name:    "valid role",
			role:    RoleEntity{ID: "123", Name: "Manager", HierarchyLevel: 50},
			wantErr: false,
		},
		{
			name:    "empty name",
			role:    RoleEntity{ID: "123", Name: "", HierarchyLevel: 50},
			wantErr: true,
		},
		{
			name:    "zero hierarchy level",
			role:    RoleEntity{ID: "123", Name: "Test", HierarchyLevel: 0},
			wantErr: true,
		},
		{
			name:    "negative hierarchy level",
			role:    RoleEntity{ID: "123", Name: "Test", HierarchyLevel: -1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRoleEntity(t *testing.T) {
	role := NewRoleEntity("Lead", "Team lead role", 25, nil, nil)
	if role.ID == "" {
		t.Error("Expected role to have an ID")
	}
	if role.Name != "Lead" {
		t.Errorf("Expected name 'Lead', got %s", role.Name)
	}
	if role.HierarchyLevel != 25 {
		t.Errorf("Expected hierarchy level 25, got %d", role.HierarchyLevel)
	}
	if role.IsSystem {
		t.Error("Expected IsSystem to be false for new roles")
	}
}
