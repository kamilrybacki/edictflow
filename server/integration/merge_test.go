//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/merge"
)

func TestMergeService_RenderManagedSection_Integration(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	// Create a team
	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Create categories
	securityCat, err := testFixtures.CreateCategory(ctx, "Security", true)
	if err != nil {
		t.Fatalf("Failed to create security category: %v", err)
	}

	testingCat, err := testFixtures.CreateCategory(ctx, "Testing", true)
	if err != nil {
		t.Fatalf("Failed to create testing category: %v", err)
	}

	// Create rules at different levels
	rule1, err := testFixtures.CreateRuleWithCategory(ctx, "No Hardcoded Secrets", team.ID, &securityCat.ID, domain.TargetLayerEnterprise, false)
	if err != nil {
		t.Fatalf("Failed to create rule1: %v", err)
	}

	rule2, err := testFixtures.CreateRuleWithCategory(ctx, "Minimum Coverage", team.ID, &testingCat.ID, domain.TargetLayerEnterprise, true)
	if err != nil {
		t.Fatalf("Failed to create rule2: %v", err)
	}

	// Render merged content
	svc := merge.NewService()
	categories := []domain.Category{securityCat, testingCat}
	rules := []domain.Rule{rule1, rule2}

	result := svc.RenderManagedSection(rules, categories)

	// Verify result
	if result == "" {
		t.Error("expected non-empty result")
	}

	if !strings.Contains(result, merge.ManagedSectionStart) {
		t.Error("expected managed section start marker")
	}

	if !strings.Contains(result, merge.ManagedSectionEnd) {
		t.Error("expected managed section end marker")
	}

	if !strings.Contains(result, "Security") {
		t.Error("expected Security category in output")
	}

	if !strings.Contains(result, "Testing") {
		t.Error("expected Testing category in output")
	}

	if !strings.Contains(result, "No Hardcoded Secrets") {
		t.Error("expected rule1 name in output")
	}

	if !strings.Contains(result, "Minimum Coverage") {
		t.Error("expected rule2 name in output")
	}

	// Rule2 is overridable, should have the tag
	if !strings.Contains(result, "(overridable)") {
		t.Error("expected overridable tag for rule2")
	}

	t.Logf("Rendered content:\n%s", result)
}

func TestMergeService_MergeWithExisting_Integration(t *testing.T) {
	svc := merge.NewService()

	existingContent := `# My Project

This is a project-specific README with some custom content.

## Custom Section

Some notes about the project.
`

	managedSection := merge.ManagedSectionStart + `

## Security

[Enterprise] **No Secrets**
Never commit API keys

` + merge.ManagedSectionEnd

	result := svc.MergeWithExisting(existingContent, managedSection)

	// Verify the original content is preserved
	if !strings.Contains(result, "# My Project") {
		t.Error("expected original content to be preserved")
	}

	if !strings.Contains(result, "Custom Section") {
		t.Error("expected custom section to be preserved")
	}

	// Verify managed section is appended
	if !strings.Contains(result, merge.ManagedSectionStart) {
		t.Error("expected managed section to be present")
	}

	if !strings.Contains(result, "No Secrets") {
		t.Error("expected rule content to be present")
	}
}

func TestMergeService_ReplaceExistingManaged_Integration(t *testing.T) {
	svc := merge.NewService()

	existingContent := `# My Project

Some intro.

` + merge.ManagedSectionStart + `

## Old Rules

[Enterprise] **Old Rule**
Old content

` + merge.ManagedSectionEnd + `

## My Custom Section

Custom content here.
`

	newManagedSection := merge.ManagedSectionStart + `

## Security

[Enterprise] **New Rule**
New content

` + merge.ManagedSectionEnd

	result := svc.MergeWithExisting(existingContent, newManagedSection)

	// Verify old managed content is replaced
	if strings.Contains(result, "Old Rule") {
		t.Error("expected old managed content to be replaced")
	}

	// Verify new managed content is present
	if !strings.Contains(result, "New Rule") {
		t.Error("expected new managed content to be present")
	}

	// Verify content before managed section is preserved
	if !strings.Contains(result, "Some intro") {
		t.Error("expected content before managed section to be preserved")
	}

	// Verify content after managed section is preserved
	if !strings.Contains(result, "My Custom Section") {
		t.Error("expected content after managed section to be preserved")
	}
}

func TestMergeService_DetectTampering_Integration(t *testing.T) {
	svc := merge.NewService()

	expected := merge.ManagedSectionStart + `

## Security

[Enterprise] **No Secrets**
Never commit API keys

` + merge.ManagedSectionEnd

	// Test no tampering
	fileContent := "before\n" + expected + "\nafter"
	if svc.DetectTampering(fileContent, expected) {
		t.Error("expected no tampering when content matches")
	}

	// Test tampering detected
	tamperedContent := "before\n" + merge.ManagedSectionStart + `

## Security

[Enterprise] **No Secrets**
MODIFIED CONTENT

` + merge.ManagedSectionEnd + "\nafter"

	if !svc.DetectTampering(tamperedContent, expected) {
		t.Error("expected tampering to be detected when content is modified")
	}

	// Test section removed
	noManagedSection := "before\nafter"
	if !svc.DetectTampering(noManagedSection, expected) {
		t.Error("expected tampering to be detected when section is removed")
	}
}
