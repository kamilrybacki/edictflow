// Package markdown provides utilities for managing CLAUDE.md files
// with managed sections that can be rendered from rules.
package markdown

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	ManagedSectionStart = "<!-- MANAGED BY EDICTFLOW - DO NOT EDIT -->"
	ManagedSectionEnd   = "<!-- END EDICTFLOW -->"
)

// Rule represents a rule to be rendered in the managed section.
// This is a minimal interface that both server domain.Rule and agent storage.CachedRule
// can be converted to.
type Rule struct {
	Name           string
	Content        string
	TargetLayer    string
	CategoryID     string
	CategoryName   string
	Overridable    bool
	PriorityWeight int
	EffectiveStart *int64
	EffectiveEnd   *int64
}

// Category represents a category for grouping rules.
type Category struct {
	ID           string
	Name         string
	DisplayOrder int
}

// IsEffective checks if a rule is currently active based on effective dates.
func (r Rule) IsEffective() bool {
	now := time.Now().Unix()
	if r.EffectiveStart != nil && now < *r.EffectiveStart {
		return false
	}
	if r.EffectiveEnd != nil && now > *r.EffectiveEnd {
		return false
	}
	return true
}

// RenderManagedSection generates the managed CLAUDE.md section from rules.
// Rules are grouped by category, sorted by DisplayOrder then name.
// Within each category, rules are sorted by priority weight (descending).
func RenderManagedSection(rules []Rule, categories []Category) string {
	if len(rules) == 0 {
		return ""
	}

	// Filter to only effective rules
	var effectiveRules []Rule
	for _, r := range rules {
		if r.IsEffective() {
			effectiveRules = append(effectiveRules, r)
		}
	}

	if len(effectiveRules) == 0 {
		return ""
	}

	// Build category lookup
	categoryMap := make(map[string]Category)
	for _, c := range categories {
		categoryMap[c.ID] = c
	}

	// Group rules by category
	rulesByCategory := make(map[string][]Rule)
	for _, r := range effectiveRules {
		catID := r.CategoryID
		rulesByCategory[catID] = append(rulesByCategory[catID], r)
	}

	// Sort categories by display order, then by name
	var sortedCatIDs []string
	for catID := range rulesByCategory {
		sortedCatIDs = append(sortedCatIDs, catID)
	}
	sort.Slice(sortedCatIDs, func(i, j int) bool {
		catI := categoryMap[sortedCatIDs[i]]
		catJ := categoryMap[sortedCatIDs[j]]
		if catI.DisplayOrder != catJ.DisplayOrder {
			return catI.DisplayOrder < catJ.DisplayOrder
		}
		return catI.Name < catJ.Name
	})

	var sections []string
	sections = append(sections, ManagedSectionStart)

	for _, catID := range sortedCatIDs {
		catRules := rulesByCategory[catID]
		cat := categoryMap[catID]

		catName := cat.Name
		if catName == "" {
			catName = "Uncategorized"
		}

		// Sort rules by priority weight within category (descending)
		sort.Slice(catRules, func(i, j int) bool {
			return catRules[i].PriorityWeight > catRules[j].PriorityWeight
		})

		sections = append(sections, fmt.Sprintf("\n## %s\n", catName))

		for _, r := range catRules {
			levelTag := fmt.Sprintf("[%s]", toTitleCase(r.TargetLayer))
			overridableTag := ""
			if r.Overridable {
				overridableTag = " (overridable)"
			}

			sections = append(sections, fmt.Sprintf("%s **%s**%s\n%s", levelTag, r.Name, overridableTag, r.Content))
		}
	}

	sections = append(sections, "\n"+ManagedSectionEnd)

	return strings.Join(sections, "\n")
}

// MergeWithExisting combines managed section with existing file content.
// If no managed section exists, it appends at the end.
// If a managed section exists, it replaces it.
func MergeWithExisting(existingContent, managedSection string) string {
	startIdx := strings.Index(existingContent, ManagedSectionStart)
	endIdx := strings.Index(existingContent, ManagedSectionEnd)

	if startIdx == -1 {
		// No existing managed section - append at end
		if existingContent != "" && !strings.HasSuffix(existingContent, "\n\n") {
			existingContent = strings.TrimRight(existingContent, "\n") + "\n\n"
		}
		return existingContent + managedSection
	}

	// Replace existing managed section
	before := existingContent[:startIdx]
	after := ""
	if endIdx != -1 {
		after = existingContent[endIdx+len(ManagedSectionEnd):]
	}

	return before + managedSection + after
}

// ExtractManualContent returns content outside the managed section.
func ExtractManualContent(content string) (before, after string) {
	startIdx := strings.Index(content, ManagedSectionStart)
	endIdx := strings.Index(content, ManagedSectionEnd)

	if startIdx == -1 {
		return content, ""
	}

	before = content[:startIdx]
	if endIdx != -1 {
		after = content[endIdx+len(ManagedSectionEnd):]
	}

	return before, after
}

// DetectTampering checks if the managed section was modified.
func DetectTampering(fileContent, expectedManaged string) bool {
	startIdx := strings.Index(fileContent, ManagedSectionStart)
	endIdx := strings.Index(fileContent, ManagedSectionEnd)

	if startIdx == -1 || endIdx == -1 {
		return expectedManaged != ""
	}

	actual := fileContent[startIdx : endIdx+len(ManagedSectionEnd)]
	return actual != expectedManaged
}

// toTitleCase capitalizes the first letter of a string.
func toTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
