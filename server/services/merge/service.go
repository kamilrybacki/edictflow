package merge

import (
	"github.com/kamilrybacki/edictflow/pkg/markdown"
	"github.com/kamilrybacki/edictflow/server/domain"
)

// Re-export constants for backward compatibility
const (
	ManagedSectionStart = markdown.ManagedSectionStart
	ManagedSectionEnd   = markdown.ManagedSectionEnd
)

// Service handles rule merging and rendering
type Service struct{}

// NewService creates a new merge service
func NewService() *Service {
	return &Service{}
}

// RenderManagedSection generates the managed CLAUDE.md section from rules
func (s *Service) RenderManagedSection(rules []domain.Rule, categories []domain.Category) string {
	// Convert domain types to shared markdown types
	mdRules := make([]markdown.Rule, 0, len(rules))
	for _, r := range rules {
		if !r.IsEffective() {
			continue
		}
		catID := ""
		if r.CategoryID != nil {
			catID = *r.CategoryID
		}
		mdRules = append(mdRules, markdown.Rule{
			Name:           r.Name,
			Content:        r.Content,
			TargetLayer:    string(r.TargetLayer),
			CategoryID:     catID,
			Overridable:    r.Overridable,
			PriorityWeight: r.PriorityWeight,
		})
	}

	mdCategories := make([]markdown.Category, 0, len(categories))
	for _, c := range categories {
		mdCategories = append(mdCategories, markdown.Category{
			ID:           c.ID,
			Name:         c.Name,
			DisplayOrder: c.DisplayOrder,
		})
	}

	return markdown.RenderManagedSection(mdRules, mdCategories)
}

// MergeWithExisting combines managed section with existing file content
func (s *Service) MergeWithExisting(existingContent, managedSection string) string {
	return markdown.MergeWithExisting(existingContent, managedSection)
}

// ExtractManualContent returns content outside the managed section
func (s *Service) ExtractManualContent(content string) (before, after string) {
	return markdown.ExtractManualContent(content)
}

// DetectTampering checks if the managed section was modified
func (s *Service) DetectTampering(fileContent, expectedManaged string) bool {
	return markdown.DetectTampering(fileContent, expectedManaged)
}
