package rules_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/services/rules"
)

func TestMatcherReturnsRulesMatchingPath(t *testing.T) {
	ruleList := []domain.Rule{
		{
			ID:       "rule-1",
			Name:     "Frontend Rule",
			Content:  "# Frontend",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypePath, Pattern: "**/frontend/**"}},
		},
		{
			ID:       "rule-2",
			Name:     "Backend Rule",
			Content:  "# Backend",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypePath, Pattern: "**/backend/**"}},
		},
	}

	matcher := rules.NewMatcher(ruleList)
	ctx := rules.MatchContext{
		ProjectPath: "/home/user/myapp/frontend/src/App.tsx",
	}

	matched := matcher.Match(ctx)

	if len(matched) != 1 {
		t.Fatalf("expected 1 matched rule, got %d", len(matched))
	}
	if matched[0].ID != "rule-1" {
		t.Errorf("expected rule-1, got %s", matched[0].ID)
	}
}

func TestMatcherReturnsRulesMatchingContext(t *testing.T) {
	ruleList := []domain.Rule{
		{
			ID:       "rule-1",
			Name:     "Node Rule",
			Content:  "# Node",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypeContext, ContextTypes: []string{"node"}}},
		},
	}

	matcher := rules.NewMatcher(ruleList)
	ctx := rules.MatchContext{
		DetectedContexts: []string{"node", "typescript"},
	}

	matched := matcher.Match(ctx)

	if len(matched) != 1 {
		t.Fatalf("expected 1 matched rule, got %d", len(matched))
	}
}

func TestMatcherSortsBySpecificityDescending(t *testing.T) {
	ruleList := []domain.Rule{
		{
			ID:       "tag-rule",
			Name:     "Tag Rule",
			Content:  "# Tag",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypeTag, Tags: []string{"frontend"}}},
		},
		{
			ID:       "path-rule",
			Name:     "Path Rule",
			Content:  "# Path",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypePath, Pattern: "**/src/**"}},
		},
	}

	matcher := rules.NewMatcher(ruleList)
	ctx := rules.MatchContext{
		ProjectPath: "/home/user/myapp/src/App.tsx",
		Tags:        []string{"frontend"},
	}

	matched := matcher.Match(ctx)

	if len(matched) != 2 {
		t.Fatalf("expected 2 matched rules, got %d", len(matched))
	}
	if matched[0].ID != "path-rule" {
		t.Errorf("expected path-rule first (highest specificity), got %s", matched[0].ID)
	}
}
