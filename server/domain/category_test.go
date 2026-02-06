package domain

import (
	"testing"
)

func TestCategory_Validate(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  bool
	}{
		{
			name: "valid system category",
			category: Category{
				Name:     "Security",
				IsSystem: true,
			},
			wantErr: false,
		},
		{
			name: "valid org category",
			category: Category{
				Name:  "Frontend Patterns",
				OrgID: stringPtr("org-123"),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			category: Category{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "name too long",
			category: Category{
				Name: string(make([]byte, 101)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.category.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Category.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
