package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TargetLayer string

const (
	TargetLayerEnterprise TargetLayer = "enterprise"
	TargetLayerGlobal     TargetLayer = "global"
	TargetLayerProject    TargetLayer = "project"
	TargetLayerLocal      TargetLayer = "local"
)

type TriggerType string

const (
	TriggerTypePath    TriggerType = "path"
	TriggerTypeContext TriggerType = "context"
	TriggerTypeTag     TriggerType = "tag"
)

type Trigger struct {
	Type         TriggerType `json:"type"`
	Pattern      string      `json:"pattern,omitempty"`
	ContextTypes []string    `json:"context_types,omitempty"`
	Tags         []string    `json:"tags,omitempty"`
}

func (t Trigger) Specificity() int {
	switch t.Type {
	case TriggerTypePath:
		return 100
	case TriggerTypeContext:
		return 50
	case TriggerTypeTag:
		return 10
	}
	return 0
}

type Rule struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Content        string      `json:"content"`
	TargetLayer    TargetLayer `json:"target_layer"`
	PriorityWeight int         `json:"priority_weight"`
	Triggers       []Trigger   `json:"triggers"`
	TeamID         string      `json:"team_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func NewRule(name string, targetLayer TargetLayer, content string, triggers []Trigger, teamID string) Rule {
	now := time.Now()
	return Rule{
		ID:             uuid.New().String(),
		Name:           name,
		Content:        content,
		TargetLayer:    targetLayer,
		PriorityWeight: 0,
		Triggers:       triggers,
		TeamID:         teamID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (r Rule) Validate() error {
	if r.Name == "" {
		return errors.New("rule name cannot be empty")
	}
	if r.Content == "" {
		return errors.New("rule content cannot be empty")
	}
	if !r.TargetLayer.IsValid() {
		return errors.New("invalid target layer")
	}
	return nil
}

func (tl TargetLayer) IsValid() bool {
	switch tl {
	case TargetLayerEnterprise, TargetLayerGlobal, TargetLayerProject, TargetLayerLocal:
		return true
	}
	return false
}

func (r Rule) MaxSpecificity() int {
	maxSpecificity := 0
	for _, trigger := range r.Triggers {
		if s := trigger.Specificity(); s > maxSpecificity {
			maxSpecificity = s
		}
	}
	return maxSpecificity
}
