package markdown

import (
	"strings"
	"testing"
	"time"
)

func TestIsEffective(t *testing.T) {
	now := time.Now().Unix()
	past := now - 3600
	future := now + 3600

	tests := []struct {
		name     string
		rule     Rule
		expected bool
	}{
		{
			name:     "no dates - always effective",
			rule:     Rule{Name: "test"},
			expected: true,
		},
		{
			name:     "start in past - effective",
			rule:     Rule{Name: "test", EffectiveStart: &past},
			expected: true,
		},
		{
			name:     "start in future - not effective",
			rule:     Rule{Name: "test", EffectiveStart: &future},
			expected: false,
		},
		{
			name:     "end in future - effective",
			rule:     Rule{Name: "test", EffectiveEnd: &future},
			expected: true,
		},
		{
			name:     "end in past - not effective",
			rule:     Rule{Name: "test", EffectiveEnd: &past},
			expected: false,
		},
		{
			name:     "valid range - effective",
			rule:     Rule{Name: "test", EffectiveStart: &past, EffectiveEnd: &future},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.IsEffective(); got != tt.expected {
				t.Errorf("IsEffective() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRenderManagedSection(t *testing.T) {
	rules := []Rule{
		{Name: "Rule A", Content: "Content A", TargetLayer: "enterprise", CategoryID: "cat1", PriorityWeight: 10, Overridable: true},
		{Name: "Rule B", Content: "Content B", TargetLayer: "user", CategoryID: "cat1", PriorityWeight: 5},
		{Name: "Rule C", Content: "Content C", TargetLayer: "project", CategoryID: "cat2", PriorityWeight: 1},
	}
	categories := []Category{
		{ID: "cat1", Name: "Category One", DisplayOrder: 1},
		{ID: "cat2", Name: "Category Two", DisplayOrder: 2},
	}

	result := RenderManagedSection(rules, categories)

	// Check section markers
	if !strings.Contains(result, ManagedSectionStart) {
		t.Error("Missing managed section start marker")
	}
	if !strings.Contains(result, ManagedSectionEnd) {
		t.Error("Missing managed section end marker")
	}

	// Check category headers appear in order
	cat1Idx := strings.Index(result, "## Category One")
	cat2Idx := strings.Index(result, "## Category Two")
	if cat1Idx == -1 || cat2Idx == -1 {
		t.Error("Missing category headers")
	}
	if cat1Idx > cat2Idx {
		t.Error("Categories not sorted by display order")
	}

	// Check rules are rendered with tags
	if !strings.Contains(result, "[Enterprise] **Rule A** (overridable)") {
		t.Error("Rule A not rendered correctly with overridable tag")
	}
	if !strings.Contains(result, "[User] **Rule B**") {
		t.Error("Rule B not rendered correctly")
	}
}

func TestRenderManagedSection_Empty(t *testing.T) {
	result := RenderManagedSection(nil, nil)
	if result != "" {
		t.Errorf("Expected empty string for no rules, got %q", result)
	}

	result = RenderManagedSection([]Rule{}, []Category{})
	if result != "" {
		t.Errorf("Expected empty string for empty rules, got %q", result)
	}
}

func TestRenderManagedSection_CategorySorting(t *testing.T) {
	rules := []Rule{
		{Name: "Rule A", Content: "A", CategoryID: "cat3"},
		{Name: "Rule B", Content: "B", CategoryID: "cat1"},
		{Name: "Rule C", Content: "C", CategoryID: "cat2"},
	}
	categories := []Category{
		{ID: "cat1", Name: "Alpha", DisplayOrder: 2},
		{ID: "cat2", Name: "Beta", DisplayOrder: 1},
		{ID: "cat3", Name: "Gamma", DisplayOrder: 2},
	}

	result := RenderManagedSection(rules, categories)

	// Beta (order 1) should come first, then Alpha (order 2, name < Gamma), then Gamma
	betaIdx := strings.Index(result, "## Beta")
	alphaIdx := strings.Index(result, "## Alpha")
	gammaIdx := strings.Index(result, "## Gamma")

	if betaIdx > alphaIdx || betaIdx > gammaIdx {
		t.Error("Beta should come first (lowest display order)")
	}
	if alphaIdx > gammaIdx {
		t.Error("Alpha should come before Gamma (same display order, alphabetically)")
	}
}

func TestMergeWithExisting(t *testing.T) {
	managedSection := ManagedSectionStart + "\nTest content\n" + ManagedSectionEnd

	t.Run("no existing section", func(t *testing.T) {
		existing := "# My Manual Content\n\nSome text."
		result := MergeWithExisting(existing, managedSection)

		if !strings.HasPrefix(result, "# My Manual Content") {
			t.Error("Original content should be preserved at start")
		}
		if !strings.HasSuffix(result, ManagedSectionEnd) {
			t.Error("Managed section should be at end")
		}
	})

	t.Run("replace existing section", func(t *testing.T) {
		existing := "Before\n" + ManagedSectionStart + "\nOld content\n" + ManagedSectionEnd + "\nAfter"
		result := MergeWithExisting(existing, managedSection)

		if !strings.HasPrefix(result, "Before\n") {
			t.Error("Content before managed section should be preserved")
		}
		if !strings.HasSuffix(result, "\nAfter") {
			t.Error("Content after managed section should be preserved")
		}
		if strings.Contains(result, "Old content") {
			t.Error("Old managed content should be replaced")
		}
		if !strings.Contains(result, "Test content") {
			t.Error("New managed content should be present")
		}
	})
}

func TestExtractManualContent(t *testing.T) {
	t.Run("no managed section", func(t *testing.T) {
		content := "Just manual content"
		before, after := ExtractManualContent(content)
		if before != content || after != "" {
			t.Error("All content should be in 'before' when no managed section")
		}
	})

	t.Run("with managed section", func(t *testing.T) {
		content := "Before\n" + ManagedSectionStart + "\nManaged\n" + ManagedSectionEnd + "\nAfter"
		before, after := ExtractManualContent(content)
		if before != "Before\n" {
			t.Errorf("Before = %q, want 'Before\\n'", before)
		}
		if after != "\nAfter" {
			t.Errorf("After = %q, want '\\nAfter'", after)
		}
	})
}

func TestDetectTampering(t *testing.T) {
	expected := ManagedSectionStart + "\nExpected\n" + ManagedSectionEnd

	t.Run("no tampering", func(t *testing.T) {
		content := "Before\n" + expected + "\nAfter"
		if DetectTampering(content, expected) {
			t.Error("Should not detect tampering when content matches")
		}
	})

	t.Run("content modified", func(t *testing.T) {
		tampered := ManagedSectionStart + "\nModified\n" + ManagedSectionEnd
		content := "Before\n" + tampered + "\nAfter"
		if !DetectTampering(content, expected) {
			t.Error("Should detect tampering when content differs")
		}
	})

	t.Run("section missing when expected", func(t *testing.T) {
		content := "Just manual content"
		if !DetectTampering(content, expected) {
			t.Error("Should detect tampering when section is missing but expected")
		}
	})

	t.Run("section missing when not expected", func(t *testing.T) {
		content := "Just manual content"
		if DetectTampering(content, "") {
			t.Error("Should not detect tampering when section is not expected")
		}
	})
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"enterprise", "Enterprise"},
		{"user", "User"},
		{"PROJECT", "PROJECT"},
		{"aBC", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := toTitleCase(tt.input); got != tt.expected {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
