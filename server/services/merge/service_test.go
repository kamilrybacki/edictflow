package merge

import (
	"strings"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
)

func strPtr(s string) *string {
	return &s
}

func TestMergeService_RenderManagedSection(t *testing.T) {
	categories := []domain.Category{
		{ID: "cat-1", Name: "Security", DisplayOrder: 1},
		{ID: "cat-2", Name: "Testing", DisplayOrder: 2},
	}

	rules := []domain.Rule{
		{
			Name:        "No Secrets",
			Content:     "Never commit API keys",
			CategoryID:  strPtr("cat-1"),
			TargetLayer: domain.TargetLayerEnterprise,
			Overridable: false,
		},
		{
			Name:        "Min Coverage",
			Content:     "Maintain 80% coverage",
			CategoryID:  strPtr("cat-2"),
			TargetLayer: domain.TargetLayerEnterprise,
			Overridable: true,
		},
	}

	svc := NewService()
	result := svc.RenderManagedSection(rules, categories)

	// Check it contains the markers
	if !strings.Contains(result, ManagedSectionStart) {
		t.Error("expected managed section start marker")
	}
	if !strings.Contains(result, ManagedSectionEnd) {
		t.Error("expected managed section end marker")
	}

	// Check it contains the category headers
	if !strings.Contains(result, "## Security") {
		t.Error("expected Security category header")
	}
	if !strings.Contains(result, "## Testing") {
		t.Error("expected Testing category header")
	}

	// Check it contains the rules
	if !strings.Contains(result, "No Secrets") {
		t.Error("expected No Secrets rule")
	}
	if !strings.Contains(result, "Min Coverage") {
		t.Error("expected Min Coverage rule")
	}

	// Check overridable tag
	if !strings.Contains(result, "(overridable)") {
		t.Error("expected overridable tag for Min Coverage")
	}

	// Check non-overridable rule doesn't have the tag
	noSecretsIdx := strings.Index(result, "No Secrets")
	overridableIdx := strings.Index(result[:noSecretsIdx+50], "(overridable)")
	if overridableIdx != -1 && overridableIdx < noSecretsIdx+len("No Secrets")+10 {
		t.Error("No Secrets should not have overridable tag")
	}
}

func TestMergeService_RenderManagedSection_Empty(t *testing.T) {
	svc := NewService()
	result := svc.RenderManagedSection(nil, nil)

	if result != "" {
		t.Errorf("expected empty string for no rules, got %q", result)
	}
}

func TestMergeService_MergeWithExisting(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name     string
		existing string
		managed  string
		want     string
	}{
		{
			name:     "no existing content",
			existing: "",
			managed:  ManagedSectionStart + "\ncontent\n" + ManagedSectionEnd,
			want:     ManagedSectionStart + "\ncontent\n" + ManagedSectionEnd,
		},
		{
			name:     "append to existing",
			existing: "# My Project\n\nSome info",
			managed:  ManagedSectionStart + "\ncontent\n" + ManagedSectionEnd,
			want:     "# My Project\n\nSome info\n\n" + ManagedSectionStart + "\ncontent\n" + ManagedSectionEnd,
		},
		{
			name:     "replace existing managed",
			existing: "before\n" + ManagedSectionStart + "\nold\n" + ManagedSectionEnd + "\nafter",
			managed:  ManagedSectionStart + "\nnew\n" + ManagedSectionEnd,
			want:     "before\n" + ManagedSectionStart + "\nnew\n" + ManagedSectionEnd + "\nafter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.MergeWithExisting(tt.existing, tt.managed)
			if got != tt.want {
				t.Errorf("MergeWithExisting() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMergeService_ExtractManualContent(t *testing.T) {
	svc := NewService()

	content := "before\n" + ManagedSectionStart + "\nmanaged\n" + ManagedSectionEnd + "\nafter"
	before, after := svc.ExtractManualContent(content)

	if before != "before\n" {
		t.Errorf("before = %q, want %q", before, "before\n")
	}
	if after != "\nafter" {
		t.Errorf("after = %q, want %q", after, "\nafter")
	}
}

func TestMergeService_DetectTampering(t *testing.T) {
	svc := NewService()

	expected := ManagedSectionStart + "\noriginal\n" + ManagedSectionEnd

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "no tampering",
			content: "before\n" + expected + "\nafter",
			want:    false,
		},
		{
			name:    "content modified",
			content: "before\n" + ManagedSectionStart + "\nmodified\n" + ManagedSectionEnd + "\nafter",
			want:    true,
		},
		{
			name:    "section removed",
			content: "before\nafter",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.DetectTampering(tt.content, expected)
			if got != tt.want {
				t.Errorf("DetectTampering() = %v, want %v", got, tt.want)
			}
		})
	}
}
