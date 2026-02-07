package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TargetLayer string

const (
	TargetLayerOrganization TargetLayer = "organization"
	TargetLayerTeam         TargetLayer = "team"
	TargetLayerProject      TargetLayer = "project"
	// Deprecated: use TargetLayerOrganization instead
	TargetLayerEnterprise TargetLayer = "enterprise"
	// Deprecated: use TargetLayerTeam instead
	TargetLayerUser TargetLayer = "user"
	// Deprecated: use TargetLayerTeam instead
	TargetLayerGlobal TargetLayer = "global"
	// Deprecated: use TargetLayerProject instead
	TargetLayerLocal TargetLayer = "local"
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
	Description           *string         `json:"description,omitempty"`
	TargetLayer           TargetLayer     `json:"target_layer"`
	CategoryID            *string         `json:"category_id,omitempty"`
	PriorityWeight        int             `json:"priority_weight"`
	Overridable           bool            `json:"overridable"`
	EffectiveStart        *time.Time      `json:"effective_start,omitempty"`
	EffectiveEnd          *time.Time      `json:"effective_end,omitempty"`
	TargetTeams           []string        `json:"target_teams,omitempty"`
	TargetUsers           []string        `json:"target_users,omitempty"`
	Tags                  []string        `json:"tags,omitempty"`
	Triggers              []Trigger       `json:"triggers"`
	TeamID                *string         `json:"team_id,omitempty"`
	Force                 bool            `json:"force"`
	Status                RuleStatus      `json:"status"`
	EnforcementMode       EnforcementMode `json:"enforcement_mode"`
	TemporaryTimeoutHours int             `json:"temporary_timeout_hours"`
	CreatedBy             *string         `json:"created_by,omitempty"`
	ApprovedBy            *string         `json:"approved_by,omitempty"`
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
		Overridable:           true,
		Triggers:              triggers,
		TeamID:                &teamID,
		Force:                 false,
		Status:                RuleStatusDraft,
		EnforcementMode:       EnforcementModeBlock,
		TemporaryTimeoutHours: 24,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func NewGlobalRule(name string, content string, force bool) Rule {
	now := time.Now()
	return Rule{
		ID:                    uuid.New().String(),
		Name:                  name,
		Content:               content,
		TargetLayer:           TargetLayerOrganization,
		PriorityWeight:        0,
		Overridable:           true,
		Triggers:              nil,
		TeamID:                nil,
		Force:                 force,
		Status:                RuleStatusDraft,
		EnforcementMode:       EnforcementModeBlock,
		TemporaryTimeoutHours: 24,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// IsGlobal returns true if this is a global rule (no team ownership)
func (r *Rule) IsGlobal() bool {
	return r.TeamID == nil || *r.TeamID == ""
}

// IsEnterprise returns true if this rule applies to all teams
func (r *Rule) IsEnterprise() bool {
	return r.TargetLayer == TargetLayerOrganization || r.TargetLayer == TargetLayerEnterprise
}

// NewLibraryRule creates a new library rule (no team ownership)
func NewLibraryRule(name string, targetLayer TargetLayer, content string, triggers []Trigger, createdBy string) Rule {
	now := time.Now()
	return Rule{
		ID:             uuid.New().String(),
		Name:           name,
		Content:        content,
		TargetLayer:    targetLayer,
		PriorityWeight: 0,
		Overridable:    true,
		Triggers:       triggers,
		Status:         RuleStatusDraft,
		CreatedBy:      &createdBy,
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
	// Global rule constraints
	if r.IsGlobal() {
		if r.TargetLayer != TargetLayerOrganization && r.TargetLayer != TargetLayerEnterprise {
			return errors.New("global rules must have organization target layer")
		}
	} else {
		// Team rule constraints
		if r.Force {
			return errors.New("force flag is only valid for global rules")
		}
	}
	return nil
}

func (tl TargetLayer) IsValid() bool {
	switch tl {
	case TargetLayerOrganization, TargetLayerTeam, TargetLayerProject,
		TargetLayerEnterprise, TargetLayerUser, TargetLayerGlobal, TargetLayerLocal:
		return true
	}
	return false
}

// ValidateOverrideConflict checks if this rule conflicts with non-overridable higher-level rules
func (r *Rule) ValidateOverrideConflict(higherRules []Rule) error {
	for _, hr := range higherRules {
		if !hr.Overridable && r.CategoryID != nil && hr.CategoryID != nil && *r.CategoryID == *hr.CategoryID {
			return fmt.Errorf("cannot create rule in category: conflicts with non-overridable %s rule '%s'", hr.TargetLayer, hr.Name)
		}
	}
	return nil
}

// IsEffective returns true if the rule is currently active based on effective dates
func (r *Rule) IsEffective() bool {
	now := time.Now()

	if r.EffectiveStart != nil && now.Before(*r.EffectiveStart) {
		return false
	}
	if r.EffectiveEnd != nil && now.After(*r.EffectiveEnd) {
		return false
	}
	return true
}

// TargetLayerPriority returns the hierarchy level (higher = more authoritative)
func (r *Rule) TargetLayerPriority() int {
	switch r.TargetLayer {
	case TargetLayerOrganization, TargetLayerEnterprise:
		return 3
	case TargetLayerTeam, TargetLayerUser, TargetLayerGlobal:
		return 2
	case TargetLayerProject, TargetLayerLocal:
		return 1
	default:
		return 0
	}
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
