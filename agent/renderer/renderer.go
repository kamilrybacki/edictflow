// agent/renderer/renderer.go
package renderer

import (
	"time"

	"github.com/kamilrybacki/edictflow/agent/storage"
	"github.com/kamilrybacki/edictflow/pkg/markdown"
)

// Re-export constants for backward compatibility
const (
	ManagedSectionStart = markdown.ManagedSectionStart
	ManagedSectionEnd   = markdown.ManagedSectionEnd
)

// Renderer generates managed content for CLAUDE.md files
type Renderer struct{}

// New creates a new Renderer
func New() *Renderer {
	return &Renderer{}
}

// RenderManagedSection generates the managed content from cached rules.
// This overload uses category names from rules for backward compatibility.
func (r *Renderer) RenderManagedSection(rules []storage.CachedRule) string {
	return r.RenderManagedSectionWithCategories(rules, nil)
}

// RenderManagedSectionWithCategories generates managed content using proper category ordering.
// When categories is nil or empty, it falls back to alphabetical sorting by category name.
func (r *Renderer) RenderManagedSectionWithCategories(rules []storage.CachedRule, categories []storage.CachedCategory) string {
	// Convert cached rules to shared markdown types
	// Filter by effective dates during conversion
	now := time.Now().Unix()
	var mdRules []markdown.Rule
	categorySet := make(map[string]bool)

	for _, rule := range rules {
		if rule.EffectiveStart != nil && now < *rule.EffectiveStart {
			continue
		}
		if rule.EffectiveEnd != nil && now > *rule.EffectiveEnd {
			continue
		}

		catID := rule.CategoryID
		catName := rule.CategoryName
		if catName == "" {
			catName = "Uncategorized"
			catID = ""
		}

		categorySet[catID] = true

		mdRules = append(mdRules, markdown.Rule{
			Name:         rule.Name,
			Content:      rule.Content,
			TargetLayer:  rule.TargetLayer,
			CategoryID:   catID,
			CategoryName: catName,
			Overridable:  rule.Overridable,
		})
	}

	// Build categories from provided categories or fall back to alphabetical
	var mdCategories []markdown.Category
	if len(categories) > 0 {
		// Use provided categories with proper display order
		for _, c := range categories {
			if categorySet[c.ID] {
				mdCategories = append(mdCategories, markdown.Category{
					ID:           c.ID,
					Name:         c.Name,
					DisplayOrder: c.DisplayOrder,
				})
			}
		}
		// Add uncategorized if needed
		if categorySet[""] {
			mdCategories = append(mdCategories, markdown.Category{
				ID:           "",
				Name:         "Uncategorized",
				DisplayOrder: 9999, // Sort last
			})
		}
	} else {
		// Fallback: build categories from rules, all with DisplayOrder 0 (alphabetical)
		for catID := range categorySet {
			catName := catID
			if catName == "" {
				catName = "Uncategorized"
			}
			mdCategories = append(mdCategories, markdown.Category{
				ID:           catID,
				Name:         catName,
				DisplayOrder: 0,
			})
		}
	}

	return markdown.RenderManagedSection(mdRules, mdCategories)
}

// MergeWithFile combines managed section with existing file content
func (r *Renderer) MergeWithFile(existing, managed string) string {
	return markdown.MergeWithExisting(existing, managed)
}

// DetectManagedSectionTampering checks if managed section was modified
func (r *Renderer) DetectManagedSectionTampering(fileContent, expectedManaged string) bool {
	return markdown.DetectTampering(fileContent, expectedManaged)
}

// ExtractManualContent returns content outside the managed section
func (r *Renderer) ExtractManualContent(content string) (before, after string) {
	return markdown.ExtractManualContent(content)
}
