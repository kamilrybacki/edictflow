package domain

import (
	"testing"
)

func TestPermission_Validate(t *testing.T) {
	tests := []struct {
		name    string
		perm    Permission
		wantErr bool
	}{
		{
			name:    "valid permission",
			perm:    Permission{ID: "123", Code: "create_rules", Category: PermissionCategoryRules},
			wantErr: false,
		},
		{
			name:    "empty code",
			perm:    Permission{ID: "123", Code: "", Category: PermissionCategoryRules},
			wantErr: true,
		},
		{
			name:    "invalid category",
			perm:    Permission{ID: "123", Code: "test", Category: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.perm.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
