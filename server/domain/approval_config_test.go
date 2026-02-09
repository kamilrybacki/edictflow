package domain

import (
	"testing"
)

func TestApprovalConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ApprovalConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  ApprovalConfig{Scope: TargetLayerTeam, RequiredPermission: "approve_global", RequiredCount: 2},
			wantErr: false,
		},
		{
			name:    "invalid scope",
			config:  ApprovalConfig{Scope: "invalid", RequiredPermission: "approve_global", RequiredCount: 2},
			wantErr: true,
		},
		{
			name:    "zero required count",
			config:  ApprovalConfig{Scope: TargetLayerTeam, RequiredPermission: "approve_global", RequiredCount: 0},
			wantErr: true,
		},
		{
			name:    "empty permission",
			config:  ApprovalConfig{Scope: TargetLayerTeam, RequiredPermission: "", RequiredCount: 2},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApprovalConfig_CanOverrideWith(t *testing.T) {
	global := ApprovalConfig{Scope: TargetLayerTeam, RequiredCount: 2}

	// Can tighten (increase)
	if !global.CanOverrideWith(3) {
		t.Error("Should allow increasing required count")
	}

	// Cannot loosen (decrease)
	if global.CanOverrideWith(1) {
		t.Error("Should not allow decreasing required count")
	}

	// Can keep same
	if !global.CanOverrideWith(2) {
		t.Error("Should allow keeping same required count")
	}
}

func TestApprovalConfig_IsGlobal(t *testing.T) {
	global := ApprovalConfig{Scope: TargetLayerTeam, TeamID: nil}
	if !global.IsGlobal() {
		t.Error("Config with nil TeamID should be global")
	}

	teamID := "team-123"
	team := ApprovalConfig{Scope: TargetLayerTeam, TeamID: &teamID}
	if team.IsGlobal() {
		t.Error("Config with TeamID should not be global")
	}
}
