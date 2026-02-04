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

type RuleStatus string

const (
	RuleStatusDraft    RuleStatus = "draft"
	RuleStatusPending  RuleStatus = "pending"
	RuleStatusApproved RuleStatus = "approved"
	RuleStatusRejected RuleStatus = "rejected"
)

func (s RuleStatus) IsValid() bool {
	switch s {
	case RuleStatusDraft, RuleStatusPending, RuleStatusApproved, RuleStatusRejected:
		return true
	}
	return false
}

type TriggerType string

const (
	TriggerTypePath    TriggerType = "path"
	TriggerTypeContext TriggerType = "context"
	TriggerTypeTag     TriggerType = "tag"
)

type EnforcementMode string

const (
	EnforcementModeBlock     EnforcementMode = "block"
	EnforcementModeTemporary EnforcementMode = "temporary"
	EnforcementModeWarning   EnforcementMode = "warning"
)

func (e EnforcementMode) IsValid() bool {
	switch e {
	case EnforcementModeBlock, EnforcementModeTemporary, EnforcementModeWarning:
		return true
	}
	return false
}

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
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	Content               string          `json:"content"`
	TargetLayer           TargetLayer     `json:"target_layer"`
	PriorityWeight        int             `json:"priority_weight"`
	Triggers              []Trigger       `json:"triggers"`
	TeamID                string          `json:"team_id"`
	Status                RuleStatus      `json:"status"`
	EnforcementMode       EnforcementMode `json:"enforcement_mode"`
	TemporaryTimeoutHours int             `json:"temporary_timeout_hours"`
	CreatedBy             *string         `json:"created_by,omitempty"`
	SubmittedAt           *time.Time      `json:"submitted_at,omitempty"`
	ApprovedAt            *time.Time      `json:"approved_at,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

func NewRule(name string, targetLayer TargetLayer, content string, triggers []Trigger, teamID string) Rule {
	now := time.Now()
	return Rule{
		ID:                    uuid.New().String(),
		Name:                  name,
		Content:               content,
		TargetLayer:           targetLayer,
		PriorityWeight:        0,
		Triggers:              triggers,
		TeamID:                teamID,
		Status:                RuleStatusDraft,
		EnforcementMode:       EnforcementModeBlock,
		TemporaryTimeoutHours: 24,
		CreatedAt:             now,
		UpdatedAt:             now,
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

func (r *Rule) CanSubmit() bool {
	return r.Status == RuleStatusDraft || r.Status == RuleStatusRejected
}

func (r *Rule) Submit() {
	r.Status = RuleStatusPending
	now := time.Now()
	r.SubmittedAt = &now
	r.UpdatedAt = now
}

func (r *Rule) Approve() {
	r.Status = RuleStatusApproved
	now := time.Now()
	r.ApprovedAt = &now
	r.UpdatedAt = now
}

func (r *Rule) Reject() {
	r.Status = RuleStatusRejected
	r.UpdatedAt = time.Now()
}

func (r *Rule) ResetToDraft() {
	r.Status = RuleStatusDraft
	r.SubmittedAt = nil
	r.ApprovedAt = nil
	r.UpdatedAt = time.Now()
}
