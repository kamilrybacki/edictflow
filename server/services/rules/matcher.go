package rules

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type MatchContext struct {
	ProjectPath      string
	DetectedContexts []string
	Tags             []string
}

type Matcher struct {
	rules []domain.Rule
}

func NewMatcher(rules []domain.Rule) *Matcher {
	return &Matcher{rules: rules}
}

func (m *Matcher) Match(ctx MatchContext) []domain.Rule {
	var matchedRules []domain.Rule

	for _, rule := range m.rules {
		if m.ruleMatches(rule, ctx) {
			matchedRules = append(matchedRules, rule)
		}
	}

	// Sort by specificity descending
	sort.Slice(matchedRules, func(i, j int) bool {
		return matchedRules[i].MaxSpecificity() > matchedRules[j].MaxSpecificity()
	})

	return matchedRules
}

func (m *Matcher) ruleMatches(rule domain.Rule, ctx MatchContext) bool {
	for _, trigger := range rule.Triggers {
		if m.triggerMatches(trigger, ctx) {
			return true
		}
	}
	return false
}

func (m *Matcher) triggerMatches(trigger domain.Trigger, ctx MatchContext) bool {
	switch trigger.Type {
	case domain.TriggerTypePath:
		return matchPath(trigger.Pattern, ctx.ProjectPath)
	case domain.TriggerTypeContext:
		return matchContext(trigger.ContextTypes, ctx.DetectedContexts)
	case domain.TriggerTypeTag:
		return matchTags(trigger.Tags, ctx.Tags)
	}
	return false
}

func matchPath(pattern, path string) bool {
	if strings.Contains(pattern, "**") {
		// Simplified ** matching: extract the middle part between ** segments
		// For patterns like **/frontend/**, extract "frontend"
		// For patterns like **/src/**, extract "src"
		parts := strings.Split(pattern, "**")
		for _, part := range parts {
			// Clean up the part by removing leading/trailing slashes
			cleaned := strings.Trim(part, "/")
			if cleaned != "" {
				// Check if the path contains this segment as a directory
				if strings.Contains(path, "/"+cleaned+"/") || strings.HasSuffix(path, "/"+cleaned) {
					return true
				}
			}
		}
		return false
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}

func matchContext(triggerContexts, detectedContexts []string) bool {
	for _, tc := range triggerContexts {
		for _, dc := range detectedContexts {
			if strings.EqualFold(tc, dc) {
				return true
			}
		}
	}
	return false
}

func matchTags(triggerTags, projectTags []string) bool {
	for _, tt := range triggerTags {
		for _, pt := range projectTags {
			if strings.EqualFold(tt, pt) {
				return true
			}
		}
	}
	return false
}
