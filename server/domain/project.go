package domain

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          string    `json:"id"`
	PathPattern string    `json:"path_pattern"`
	Tags        []string  `json:"tags"`
	TeamID      string    `json:"team_id"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewProject(pathPattern string, tags []string, teamID string) Project {
	now := time.Now()
	return Project{
		ID:          uuid.New().String(),
		PathPattern: pathPattern,
		Tags:        tags,
		TeamID:      teamID,
		LastSeenAt:  now,
		CreatedAt:   now,
	}
}

func (p Project) MatchesPath(path string) bool {
	// For glob patterns with **, use globMatch since filepath.Match doesn't support **
	// Real implementation would use github.com/bmatcuk/doublestar
	if strings.Contains(p.PathPattern, "**") {
		return globMatch(p.PathPattern, path)
	}
	matched, _ := filepath.Match(p.PathPattern, path)
	return matched
}

func globMatch(pattern, path string) bool {
	// Simplified glob matching for ** patterns
	// In production, use github.com/bmatcuk/doublestar
	if len(pattern) >= 2 && pattern[:2] == "**" {
		suffix := pattern[2:]
		if len(suffix) > 0 && suffix[0] == '/' {
			suffix = suffix[1:]
		}
		// Handle patterns like **/frontend/**
		// Check if path contains the middle segment
		if len(suffix) >= 2 && suffix[len(suffix)-2:] == "**" {
			// Pattern like **/frontend/**
			middle := suffix[:len(suffix)-3] // Remove /** at end
			if len(middle) > 0 && middle[len(middle)-1] == '/' {
				middle = middle[:len(middle)-1]
			}
			return contains(path, "/"+middle+"/")
		}
		// Check if path contains the suffix pattern
		matched, _ := filepath.Match("*"+suffix+"*", path)
		return matched
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
