package domain_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewProjectCreatesValidProject(t *testing.T) {
	project := domain.NewProject("~/projects/myapp", []string{"frontend", "react"}, "team-123")

	if project.PathPattern != "~/projects/myapp" {
		t.Errorf("expected path pattern '~/projects/myapp', got '%s'", project.PathPattern)
	}
	if len(project.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(project.Tags))
	}
}

func TestProjectMatchesPathPattern(t *testing.T) {
	project := domain.NewProject("**/frontend/**", []string{}, "team-123")

	if !project.MatchesPath("/home/user/projects/myapp/frontend/src/App.tsx") {
		t.Error("expected path to match pattern")
	}
	if project.MatchesPath("/home/user/projects/myapp/backend/main.go") {
		t.Error("expected path NOT to match pattern")
	}
}
