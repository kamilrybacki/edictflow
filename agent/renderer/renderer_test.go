package renderer

import (
	"strings"
	"testing"

	"github.com/kamilrybacki/edictflow/agent/storage"
)

func TestRenderer_RenderManagedSection(t *testing.T) {
	rules := []storage.CachedRule{
		{
			Name:         "No Secrets",
			Content:      "Never commit API keys",
			CategoryName: "Security",
			TargetLayer:  "enterprise",
			Overridable:  false,
		},
		{
			Name:         "Min Coverage",
			Content:      "Maintain 80% coverage",
			CategoryName: "Testing",
			TargetLayer:  "enterprise",
			Overridable:  true,
		},
	}

	r := New()
	result := r.RenderManagedSection(rules)

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !strings.Contains(result, ManagedSectionStart) {
		t.Error("expected managed section start marker")
	}

	if !strings.Contains(result, ManagedSectionEnd) {
		t.Error("expected managed section end marker")
	}

	if !strings.Contains(result, "No Secrets") {
		t.Error("expected rule name in output")
	}

	if !strings.Contains(result, "(overridable)") {
		t.Error("expected overridable tag for Min Coverage")
	}
}

func TestRenderer_RenderManagedSection_Empty(t *testing.T) {
	r := New()
	result := r.RenderManagedSection(nil)

	if result != "" {
		t.Errorf("expected empty string for no rules, got %q", result)
	}
}

func TestRenderer_MergeWithFile(t *testing.T) {
	r := New()

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
			got := r.MergeWithFile(tt.existing, tt.managed)
			if got != tt.want {
				t.Errorf("MergeWithFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderer_DetectManagedSectionTampering(t *testing.T) {
	r := New()

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
			got := r.DetectManagedSectionTampering(tt.content, expected)
			if got != tt.want {
				t.Errorf("DetectManagedSectionTampering() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderer_EffectiveDateFiltering(t *testing.T) {
	r := New()

	past := int64(1000)
	future := int64(9999999999)

	rules := []storage.CachedRule{
		{
			Name:         "Active Rule",
			Content:      "active content",
			CategoryName: "General",
			TargetLayer:  "project",
		},
		{
			Name:           "Future Rule",
			Content:        "future content",
			CategoryName:   "General",
			TargetLayer:    "project",
			EffectiveStart: &future,
		},
		{
			Name:         "Expired Rule",
			Content:      "expired content",
			CategoryName: "General",
			TargetLayer:  "project",
			EffectiveEnd: &past,
		},
	}

	result := r.RenderManagedSection(rules)

	if !strings.Contains(result, "Active Rule") {
		t.Error("expected Active Rule to be included")
	}

	if strings.Contains(result, "Future Rule") {
		t.Error("expected Future Rule to be excluded")
	}

	if strings.Contains(result, "Expired Rule") {
		t.Error("expected Expired Rule to be excluded")
	}
}
