# Rule Library Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor the rules system so all rules live in a central library and are attached to teams with per-attachment enforcement settings.

**Architecture:** Introduce a `RuleAttachment` entity that links library rules to teams. Rules lose their `team_id`, `enforcement_mode`, and `force` fields. Enterprise rules auto-attach to all teams on approval. Attachments require separate approval.

**Tech Stack:** Go (server), PostgreSQL (migrations), React/TypeScript (frontend), chi router (API)

---

## Phase 1: Domain Layer

### Task 1: Create RuleAttachment Domain Entity

**Files:**
- Create: `server/domain/rule_attachment.go`
- Create: `server/domain/rule_attachment_test.go`

**Step 1: Write the failing test**

```go
// server/domain/rule_attachment_test.go
package domain_test

import (
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
)

func TestNewRuleAttachment(t *testing.T) {
	attachment := domain.NewRuleAttachment("rule-123", "team-456", domain.EnforcementModeBlock, "user-789")

	if attachment.RuleID != "rule-123" {
		t.Errorf("expected RuleID 'rule-123', got '%s'", attachment.RuleID)
	}
	if attachment.TeamID != "team-456" {
		t.Errorf("expected TeamID 'team-456', got '%s'", attachment.TeamID)
	}
	if attachment.EnforcementMode != domain.EnforcementModeBlock {
		t.Errorf("expected EnforcementMode 'block', got '%s'", attachment.EnforcementMode)
	}
	if attachment.Status != domain.AttachmentStatusPending {
		t.Errorf("expected Status 'pending', got '%s'", attachment.Status)
	}
	if attachment.RequestedBy != "user-789" {
		t.Errorf("expected RequestedBy 'user-789', got '%s'", attachment.RequestedBy)
	}
}

func TestRuleAttachmentValidate(t *testing.T) {
	tests := []struct {
		name       string
		attachment domain.RuleAttachment
		wantErr    bool
	}{
		{
			name: "valid attachment",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				RuleID:          "rule-1",
				TeamID:          "team-1",
				EnforcementMode: domain.EnforcementModeBlock,
				RequestedBy:     "user-1",
			},
			wantErr: false,
		},
		{
			name: "missing rule ID",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				TeamID:          "team-1",
				EnforcementMode: domain.EnforcementModeBlock,
				RequestedBy:     "user-1",
			},
			wantErr: true,
		},
		{
			name: "missing team ID",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				RuleID:          "rule-1",
				EnforcementMode: domain.EnforcementModeBlock,
				RequestedBy:     "user-1",
			},
			wantErr: true,
		},
		{
			name: "invalid enforcement mode",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				RuleID:          "rule-1",
				TeamID:          "team-1",
				EnforcementMode: "invalid",
				RequestedBy:     "user-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attachment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleAttachment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleAttachmentApprove(t *testing.T) {
	attachment := domain.NewRuleAttachment("rule-1", "team-1", domain.EnforcementModeBlock, "user-1")

	attachment.Approve("admin-1")

	if attachment.Status != domain.AttachmentStatusApproved {
		t.Errorf("expected status 'approved', got '%s'", attachment.Status)
	}
	if attachment.ApprovedBy == nil || *attachment.ApprovedBy != "admin-1" {
		t.Error("expected ApprovedBy to be set")
	}
	if attachment.ApprovedAt == nil {
		t.Error("expected ApprovedAt to be set")
	}
}

func TestRuleAttachmentReject(t *testing.T) {
	attachment := domain.NewRuleAttachment("rule-1", "team-1", domain.EnforcementModeBlock, "user-1")

	attachment.Reject()

	if attachment.Status != domain.AttachmentStatusRejected {
		t.Errorf("expected status 'rejected', got '%s'", attachment.Status)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain -run TestNewRuleAttachment -v`
Expected: FAIL with "undefined: domain.RuleAttachment"

**Step 3: Write minimal implementation**

```go
// server/domain/rule_attachment.go
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AttachmentStatus string

const (
	AttachmentStatusPending  AttachmentStatus = "pending"
	AttachmentStatusApproved AttachmentStatus = "approved"
	AttachmentStatusRejected AttachmentStatus = "rejected"
)

func (s AttachmentStatus) IsValid() bool {
	switch s {
	case AttachmentStatusPending, AttachmentStatusApproved, AttachmentStatusRejected:
		return true
	}
	return false
}

type RuleAttachment struct {
	ID                    string           `json:"id"`
	RuleID                string           `json:"rule_id"`
	TeamID                string           `json:"team_id"`
	EnforcementMode       EnforcementMode  `json:"enforcement_mode"`
	TemporaryTimeoutHours int              `json:"temporary_timeout_hours"`
	Status                AttachmentStatus `json:"status"`
	RequestedBy           string           `json:"requested_by"`
	ApprovedBy            *string          `json:"approved_by,omitempty"`
	CreatedAt             time.Time        `json:"created_at"`
	ApprovedAt            *time.Time       `json:"approved_at,omitempty"`
}

func NewRuleAttachment(ruleID, teamID string, enforcementMode EnforcementMode, requestedBy string) RuleAttachment {
	return RuleAttachment{
		ID:                    uuid.New().String(),
		RuleID:                ruleID,
		TeamID:                teamID,
		EnforcementMode:       enforcementMode,
		TemporaryTimeoutHours: 24,
		Status:                AttachmentStatusPending,
		RequestedBy:           requestedBy,
		CreatedAt:             time.Now(),
	}
}

func NewApprovedAttachment(ruleID, teamID string, enforcementMode EnforcementMode, approvedBy string) RuleAttachment {
	now := time.Now()
	return RuleAttachment{
		ID:                    uuid.New().String(),
		RuleID:                ruleID,
		TeamID:                teamID,
		EnforcementMode:       enforcementMode,
		TemporaryTimeoutHours: 24,
		Status:                AttachmentStatusApproved,
		RequestedBy:           approvedBy,
		ApprovedBy:            &approvedBy,
		CreatedAt:             now,
		ApprovedAt:            &now,
	}
}

func (a RuleAttachment) Validate() error {
	if a.RuleID == "" {
		return errors.New("rule ID cannot be empty")
	}
	if a.TeamID == "" {
		return errors.New("team ID cannot be empty")
	}
	if !a.EnforcementMode.IsValid() {
		return errors.New("invalid enforcement mode")
	}
	return nil
}

func (a *RuleAttachment) Approve(approvedBy string) {
	a.Status = AttachmentStatusApproved
	a.ApprovedBy = &approvedBy
	now := time.Now()
	a.ApprovedAt = &now
}

func (a *RuleAttachment) Reject() {
	a.Status = AttachmentStatusRejected
}

func (a *RuleAttachment) UpdateEnforcement(mode EnforcementMode, timeoutHours int) {
	a.EnforcementMode = mode
	if mode == EnforcementModeTemporary {
		a.TemporaryTimeoutHours = timeoutHours
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain -run TestRuleAttachment -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/rule_attachment.go server/domain/rule_attachment_test.go
git commit -m "feat(domain): add RuleAttachment entity for library-to-team bindings"
```

---

### Task 2: Update Rule Entity - Remove Team-Specific Fields

**Files:**
- Modify: `server/domain/rule.go`
- Modify: `server/domain/rule_test.go`

**Step 1: Write failing tests for new behavior**

Add to `server/domain/rule_test.go`:

```go
func TestNewLibraryRule(t *testing.T) {
	triggers := []domain.Trigger{
		{Type: domain.TriggerTypePath, Pattern: "*.go"},
	}
	rule := domain.NewLibraryRule("Go Standards", domain.TargetLayerTeam, "Use gofmt.", triggers, "user-123")

	if rule.Name != "Go Standards" {
		t.Errorf("expected name 'Go Standards', got '%s'", rule.Name)
	}
	if rule.CreatedBy == nil || *rule.CreatedBy != "user-123" {
		t.Error("expected CreatedBy to be set")
	}
	if rule.Status != domain.RuleStatusDraft {
		t.Errorf("expected status 'draft', got '%s'", rule.Status)
	}
}

func TestRule_IsEnterprise(t *testing.T) {
	tests := []struct {
		name        string
		targetLayer domain.TargetLayer
		want        bool
	}{
		{"organization layer", domain.TargetLayerOrganization, true},
		{"enterprise layer", domain.TargetLayerEnterprise, true},
		{"team layer", domain.TargetLayerTeam, false},
		{"project layer", domain.TargetLayerProject, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := domain.Rule{TargetLayer: tt.targetLayer}
			if got := rule.IsEnterprise(); got != tt.want {
				t.Errorf("Rule.IsEnterprise() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain -run TestNewLibraryRule -v`
Expected: FAIL with "undefined: domain.NewLibraryRule"

**Step 3: Modify Rule struct and add new constructor**

In `server/domain/rule.go`, add:

```go
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

// IsEnterprise returns true if this rule applies to all teams
func (r *Rule) IsEnterprise() bool {
	return r.TargetLayer == TargetLayerOrganization || r.TargetLayer == TargetLayerEnterprise
}
```

**Step 4: Run tests to verify they pass**

Run: `cd server && go test ./domain -run "TestNewLibraryRule|TestRule_IsEnterprise" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/rule.go server/domain/rule_test.go
git commit -m "feat(domain): add NewLibraryRule constructor and IsEnterprise method"
```

---

## Phase 2: Database Layer

### Task 3: Create Migration for rule_attachments Table

**Files:**
- Create: `server/migrations/000009_rule_attachments.up.sql`
- Create: `server/migrations/000009_rule_attachments.down.sql`

**Step 1: Write the up migration**

```sql
-- server/migrations/000009_rule_attachments.up.sql
-- Rule attachments: link library rules to teams with enforcement settings

CREATE TABLE rule_attachments (
    id UUID PRIMARY KEY,
    rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    enforcement_mode TEXT NOT NULL DEFAULT 'block' CHECK (enforcement_mode IN ('block', 'temporary', 'warning')),
    temporary_timeout_hours INTEGER NOT NULL DEFAULT 24,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(rule_id, team_id)
);

CREATE INDEX idx_rule_attachments_rule_id ON rule_attachments(rule_id);
CREATE INDEX idx_rule_attachments_team_id ON rule_attachments(team_id);
CREATE INDEX idx_rule_attachments_status ON rule_attachments(status);

-- Add approved_by column to rules table
ALTER TABLE rules ADD COLUMN approved_by UUID REFERENCES users(id) ON DELETE SET NULL;
```

**Step 2: Write the down migration**

```sql
-- server/migrations/000009_rule_attachments.down.sql
DROP TABLE IF EXISTS rule_attachments;
ALTER TABLE rules DROP COLUMN IF EXISTS approved_by;
```

**Step 3: Verify migrations are syntactically correct**

Run: `cd server && cat migrations/000009_rule_attachments.up.sql`
Expected: No errors, file contents displayed

**Step 4: Commit**

```bash
git add server/migrations/000009_rule_attachments.up.sql server/migrations/000009_rule_attachments.down.sql
git commit -m "feat(db): add rule_attachments table and approved_by column"
```

---

### Task 4: Create RuleAttachment Database Adapter

**Files:**
- Create: `server/adapters/postgres/rule_attachment_db.go`

**Step 1: Write the database adapter**

```go
// server/adapters/postgres/rule_attachment_db.go
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrAttachmentNotFound = errors.New("attachment not found")
var ErrAttachmentExists = errors.New("attachment already exists for this rule and team")

type RuleAttachmentDB struct {
	pool *pgxpool.Pool
}

func NewRuleAttachmentDB(pool *pgxpool.Pool) *RuleAttachmentDB {
	return &RuleAttachmentDB{pool: pool}
}

func (db *RuleAttachmentDB) Create(ctx context.Context, attachment domain.RuleAttachment) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO rule_attachments (
			id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, attachment.ID, attachment.RuleID, attachment.TeamID, attachment.EnforcementMode,
		attachment.TemporaryTimeoutHours, attachment.Status, attachment.RequestedBy,
		attachment.ApprovedBy, attachment.CreatedAt, attachment.ApprovedAt)

	if err != nil && err.Error() == "ERROR: duplicate key value violates unique constraint \"rule_attachments_rule_id_team_id_key\" (SQLSTATE 23505)" {
		return ErrAttachmentExists
	}
	return err
}

func (db *RuleAttachmentDB) GetByID(ctx context.Context, id string) (domain.RuleAttachment, error) {
	var att domain.RuleAttachment
	err := db.pool.QueryRow(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE id = $1
	`, id).Scan(
		&att.ID, &att.RuleID, &att.TeamID, &att.EnforcementMode, &att.TemporaryTimeoutHours,
		&att.Status, &att.RequestedBy, &att.ApprovedBy, &att.CreatedAt, &att.ApprovedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RuleAttachment{}, ErrAttachmentNotFound
	}
	return att, err
}

func (db *RuleAttachmentDB) GetByRuleAndTeam(ctx context.Context, ruleID, teamID string) (domain.RuleAttachment, error) {
	var att domain.RuleAttachment
	err := db.pool.QueryRow(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE rule_id = $1 AND team_id = $2
	`, ruleID, teamID).Scan(
		&att.ID, &att.RuleID, &att.TeamID, &att.EnforcementMode, &att.TemporaryTimeoutHours,
		&att.Status, &att.RequestedBy, &att.ApprovedBy, &att.CreatedAt, &att.ApprovedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RuleAttachment{}, ErrAttachmentNotFound
	}
	return att, err
}

func (db *RuleAttachmentDB) ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE team_id = $1
		ORDER BY created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanAttachments(rows)
}

func (db *RuleAttachmentDB) ListByRule(ctx context.Context, ruleID string) ([]domain.RuleAttachment, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE rule_id = $1
		ORDER BY created_at DESC
	`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanAttachments(rows)
}

func (db *RuleAttachmentDB) ListByStatus(ctx context.Context, status domain.AttachmentStatus) ([]domain.RuleAttachment, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE status = $1
		ORDER BY created_at ASC
	`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanAttachments(rows)
}

func (db *RuleAttachmentDB) Update(ctx context.Context, attachment domain.RuleAttachment) error {
	result, err := db.pool.Exec(ctx, `
		UPDATE rule_attachments
		SET enforcement_mode = $2, temporary_timeout_hours = $3, status = $4,
			approved_by = $5, approved_at = $6
		WHERE id = $1
	`, attachment.ID, attachment.EnforcementMode, attachment.TemporaryTimeoutHours,
		attachment.Status, attachment.ApprovedBy, attachment.ApprovedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAttachmentNotFound
	}
	return nil
}

func (db *RuleAttachmentDB) Delete(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `DELETE FROM rule_attachments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAttachmentNotFound
	}
	return nil
}

func (db *RuleAttachmentDB) scanAttachments(rows pgx.Rows) ([]domain.RuleAttachment, error) {
	var attachments []domain.RuleAttachment
	for rows.Next() {
		var att domain.RuleAttachment
		if err := rows.Scan(
			&att.ID, &att.RuleID, &att.TeamID, &att.EnforcementMode, &att.TemporaryTimeoutHours,
			&att.Status, &att.RequestedBy, &att.ApprovedBy, &att.CreatedAt, &att.ApprovedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}
	return attachments, rows.Err()
}
```

**Step 2: Verify compilation**

Run: `cd server && go build ./adapters/postgres`
Expected: No errors

**Step 3: Commit**

```bash
git add server/adapters/postgres/rule_attachment_db.go
git commit -m "feat(db): add RuleAttachmentDB adapter for PostgreSQL"
```

---

## Phase 3: Service Layer

### Task 5: Create Attachments Service

**Files:**
- Create: `server/services/attachments/service.go`
- Create: `server/services/attachments/service_test.go`

**Step 1: Write the failing test**

```go
// server/services/attachments/service_test.go
package attachments_test

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/attachments"
)

type mockAttachmentDB struct {
	attachments map[string]domain.RuleAttachment
}

func newMockDB() *mockAttachmentDB {
	return &mockAttachmentDB{attachments: make(map[string]domain.RuleAttachment)}
}

func (m *mockAttachmentDB) Create(ctx context.Context, att domain.RuleAttachment) error {
	m.attachments[att.ID] = att
	return nil
}

func (m *mockAttachmentDB) GetByID(ctx context.Context, id string) (domain.RuleAttachment, error) {
	if att, ok := m.attachments[id]; ok {
		return att, nil
	}
	return domain.RuleAttachment{}, attachments.ErrNotFound
}

func (m *mockAttachmentDB) ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error) {
	var result []domain.RuleAttachment
	for _, att := range m.attachments {
		if att.TeamID == teamID {
			result = append(result, att)
		}
	}
	return result, nil
}

func (m *mockAttachmentDB) Update(ctx context.Context, att domain.RuleAttachment) error {
	m.attachments[att.ID] = att
	return nil
}

func (m *mockAttachmentDB) Delete(ctx context.Context, id string) error {
	delete(m.attachments, id)
	return nil
}

type mockRuleDB struct{}

func (m *mockRuleDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	return domain.Rule{
		ID:          id,
		Status:      domain.RuleStatusApproved,
		TargetLayer: domain.TargetLayerTeam,
	}, nil
}

type mockTeamDB struct{}

func (m *mockTeamDB) ListAllTeams(ctx context.Context) ([]domain.Team, error) {
	return []domain.Team{
		{ID: "team-1"},
		{ID: "team-2"},
	}, nil
}

func TestService_RequestAttachment(t *testing.T) {
	db := newMockDB()
	svc := attachments.NewService(db, &mockRuleDB{}, &mockTeamDB{})

	att, err := svc.RequestAttachment(context.Background(), attachments.AttachRequest{
		RuleID:          "rule-1",
		TeamID:          "team-1",
		EnforcementMode: domain.EnforcementModeBlock,
		RequestedBy:     "user-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if att.Status != domain.AttachmentStatusPending {
		t.Errorf("expected pending status, got %s", att.Status)
	}
}

func TestService_ApproveAttachment(t *testing.T) {
	db := newMockDB()
	svc := attachments.NewService(db, &mockRuleDB{}, &mockTeamDB{})

	att, _ := svc.RequestAttachment(context.Background(), attachments.AttachRequest{
		RuleID:          "rule-1",
		TeamID:          "team-1",
		EnforcementMode: domain.EnforcementModeBlock,
		RequestedBy:     "user-1",
	})

	approved, err := svc.ApproveAttachment(context.Background(), att.ID, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved.Status != domain.AttachmentStatusApproved {
		t.Errorf("expected approved status, got %s", approved.Status)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./services/attachments -v`
Expected: FAIL (package doesn't exist)

**Step 3: Write the service**

```go
// server/services/attachments/service.go
package attachments

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrNotFound = errors.New("attachment not found")
var ErrRuleNotApproved = errors.New("rule must be approved before attaching")
var ErrAlreadyAttached = errors.New("rule is already attached to this team")

type DB interface {
	Create(ctx context.Context, attachment domain.RuleAttachment) error
	GetByID(ctx context.Context, id string) (domain.RuleAttachment, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error)
	Update(ctx context.Context, attachment domain.RuleAttachment) error
	Delete(ctx context.Context, id string) error
}

type RuleDB interface {
	GetRule(ctx context.Context, id string) (domain.Rule, error)
}

type TeamDB interface {
	ListAllTeams(ctx context.Context) ([]domain.Team, error)
}

type Service struct {
	db     DB
	ruleDB RuleDB
	teamDB TeamDB
}

func NewService(db DB, ruleDB RuleDB, teamDB TeamDB) *Service {
	return &Service{db: db, ruleDB: ruleDB, teamDB: teamDB}
}

type AttachRequest struct {
	RuleID          string
	TeamID          string
	EnforcementMode domain.EnforcementMode
	TimeoutHours    int
	RequestedBy     string
}

func (s *Service) RequestAttachment(ctx context.Context, req AttachRequest) (domain.RuleAttachment, error) {
	// Verify rule exists and is approved
	rule, err := s.ruleDB.GetRule(ctx, req.RuleID)
	if err != nil {
		return domain.RuleAttachment{}, err
	}
	if rule.Status != domain.RuleStatusApproved {
		return domain.RuleAttachment{}, ErrRuleNotApproved
	}

	attachment := domain.NewRuleAttachment(req.RuleID, req.TeamID, req.EnforcementMode, req.RequestedBy)
	if req.TimeoutHours > 0 {
		attachment.TemporaryTimeoutHours = req.TimeoutHours
	}

	if err := s.db.Create(ctx, attachment); err != nil {
		return domain.RuleAttachment{}, err
	}
	return attachment, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.RuleAttachment, error) {
	return s.db.GetByID(ctx, id)
}

func (s *Service) ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error) {
	return s.db.ListByTeam(ctx, teamID)
}

func (s *Service) ApproveAttachment(ctx context.Context, id, approvedBy string) (domain.RuleAttachment, error) {
	att, err := s.db.GetByID(ctx, id)
	if err != nil {
		return domain.RuleAttachment{}, err
	}

	att.Approve(approvedBy)
	if err := s.db.Update(ctx, att); err != nil {
		return domain.RuleAttachment{}, err
	}
	return att, nil
}

func (s *Service) RejectAttachment(ctx context.Context, id string) (domain.RuleAttachment, error) {
	att, err := s.db.GetByID(ctx, id)
	if err != nil {
		return domain.RuleAttachment{}, err
	}

	att.Reject()
	if err := s.db.Update(ctx, att); err != nil {
		return domain.RuleAttachment{}, err
	}
	return att, nil
}

func (s *Service) UpdateEnforcement(ctx context.Context, id string, mode domain.EnforcementMode, timeoutHours int) (domain.RuleAttachment, error) {
	att, err := s.db.GetByID(ctx, id)
	if err != nil {
		return domain.RuleAttachment{}, err
	}

	att.UpdateEnforcement(mode, timeoutHours)
	if err := s.db.Update(ctx, att); err != nil {
		return domain.RuleAttachment{}, err
	}
	return att, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.db.Delete(ctx, id)
}

// AutoAttachEnterpriseRule creates approved attachments for all teams
func (s *Service) AutoAttachEnterpriseRule(ctx context.Context, ruleID, approvedBy string) error {
	teams, err := s.teamDB.ListAllTeams(ctx)
	if err != nil {
		return err
	}

	for _, team := range teams {
		att := domain.NewApprovedAttachment(ruleID, team.ID, domain.EnforcementModeBlock, approvedBy)
		if err := s.db.Create(ctx, att); err != nil {
			// Ignore duplicate errors (rule might already be attached)
			continue
		}
	}
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd server && go test ./services/attachments -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/services/attachments/service.go server/services/attachments/service_test.go
git commit -m "feat(service): add attachments service for rule-team bindings"
```

---

### Task 6: Create Library Service

**Files:**
- Create: `server/services/library/service.go`
- Create: `server/services/library/service_test.go`

**Step 1: Write the failing test**

```go
// server/services/library/service_test.go
package library_test

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/library"
)

type mockRuleDB struct {
	rules map[string]domain.Rule
}

func newMockRuleDB() *mockRuleDB {
	return &mockRuleDB{rules: make(map[string]domain.Rule)}
}

func (m *mockRuleDB) CreateRule(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockRuleDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	if r, ok := m.rules[id]; ok {
		return r, nil
	}
	return domain.Rule{}, library.ErrRuleNotFound
}

func (m *mockRuleDB) ListAllRules(ctx context.Context) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.rules {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockRuleDB) UpdateRule(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockRuleDB) DeleteRule(ctx context.Context, id string) error {
	delete(m.rules, id)
	return nil
}

func (m *mockRuleDB) UpdateStatus(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

type mockAttachmentService struct {
	called bool
}

func (m *mockAttachmentService) AutoAttachEnterpriseRule(ctx context.Context, ruleID, approvedBy string) error {
	m.called = true
	return nil
}

func TestLibraryService_Create(t *testing.T) {
	db := newMockRuleDB()
	svc := library.NewService(db, nil)

	rule, err := svc.Create(context.Background(), library.CreateRequest{
		Name:        "Test Rule",
		Content:     "Test content",
		TargetLayer: domain.TargetLayerTeam,
		CreatedBy:   "user-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.Name != "Test Rule" {
		t.Errorf("expected name 'Test Rule', got '%s'", rule.Name)
	}
	if rule.Status != domain.RuleStatusDraft {
		t.Errorf("expected draft status, got '%s'", rule.Status)
	}
}

func TestLibraryService_ApproveEnterpriseRule(t *testing.T) {
	db := newMockRuleDB()
	attSvc := &mockAttachmentService{}
	svc := library.NewService(db, attSvc)

	// Create and submit enterprise rule
	rule, _ := svc.Create(context.Background(), library.CreateRequest{
		Name:        "Enterprise Policy",
		Content:     "All teams must...",
		TargetLayer: domain.TargetLayerOrganization,
		CreatedBy:   "user-1",
	})
	svc.Submit(context.Background(), rule.ID)

	// Approve
	approved, err := svc.Approve(context.Background(), rule.ID, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved.Status != domain.RuleStatusApproved {
		t.Errorf("expected approved status, got '%s'", approved.Status)
	}
	if !attSvc.called {
		t.Error("expected AutoAttachEnterpriseRule to be called")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./services/library -v`
Expected: FAIL (package doesn't exist)

**Step 3: Write the library service**

```go
// server/services/library/service.go
package library

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrRuleNotFound = errors.New("rule not found")
var ErrInvalidStatus = errors.New("rule is not in a valid status for this operation")

type DB interface {
	CreateRule(ctx context.Context, rule domain.Rule) error
	GetRule(ctx context.Context, id string) (domain.Rule, error)
	ListAllRules(ctx context.Context) ([]domain.Rule, error)
	UpdateRule(ctx context.Context, rule domain.Rule) error
	DeleteRule(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, rule domain.Rule) error
}

type AttachmentService interface {
	AutoAttachEnterpriseRule(ctx context.Context, ruleID, approvedBy string) error
}

type Service struct {
	db            DB
	attachmentSvc AttachmentService
}

func NewService(db DB, attachmentSvc AttachmentService) *Service {
	return &Service{db: db, attachmentSvc: attachmentSvc}
}

type CreateRequest struct {
	Name           string
	Content        string
	Description    string
	TargetLayer    domain.TargetLayer
	CategoryID     string
	PriorityWeight int
	Overridable    bool
	Tags           []string
	Triggers       []domain.Trigger
	CreatedBy      string
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (domain.Rule, error) {
	rule := domain.NewLibraryRule(req.Name, req.TargetLayer, req.Content, req.Triggers, req.CreatedBy)

	if req.Description != "" {
		rule.Description = &req.Description
	}
	if req.CategoryID != "" {
		rule.CategoryID = &req.CategoryID
	}
	rule.PriorityWeight = req.PriorityWeight
	rule.Overridable = req.Overridable
	rule.Tags = req.Tags

	if err := rule.Validate(); err != nil {
		return domain.Rule{}, err
	}
	if err := s.db.CreateRule(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Rule, error) {
	return s.db.GetRule(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Rule, error) {
	return s.db.ListAllRules(ctx)
}

func (s *Service) Update(ctx context.Context, rule domain.Rule) error {
	existing, err := s.db.GetRule(ctx, rule.ID)
	if err != nil {
		return err
	}
	// Only draft or rejected rules can be edited
	if existing.Status != domain.RuleStatusDraft && existing.Status != domain.RuleStatusRejected {
		return ErrInvalidStatus
	}
	return s.db.UpdateRule(ctx, rule)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return err
	}
	// Only draft rules can be deleted
	if rule.Status != domain.RuleStatusDraft {
		return ErrInvalidStatus
	}
	return s.db.DeleteRule(ctx, id)
}

func (s *Service) Submit(ctx context.Context, id string) (domain.Rule, error) {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}
	if !rule.CanSubmit() {
		return domain.Rule{}, ErrInvalidStatus
	}

	rule.Submit()
	if err := s.db.UpdateStatus(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
}

func (s *Service) Approve(ctx context.Context, id, approvedBy string) (domain.Rule, error) {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}
	if rule.Status != domain.RuleStatusPending {
		return domain.Rule{}, ErrInvalidStatus
	}

	rule.Approve()
	rule.ApprovedBy = &approvedBy
	if err := s.db.UpdateStatus(ctx, rule); err != nil {
		return domain.Rule{}, err
	}

	// Auto-attach enterprise rules to all teams
	if rule.IsEnterprise() && s.attachmentSvc != nil {
		if err := s.attachmentSvc.AutoAttachEnterpriseRule(ctx, rule.ID, approvedBy); err != nil {
			// Log but don't fail - rule is still approved
			// TODO: Add proper logging
		}
	}

	return rule, nil
}

func (s *Service) Reject(ctx context.Context, id string) (domain.Rule, error) {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}
	if rule.Status != domain.RuleStatusPending {
		return domain.Rule{}, ErrInvalidStatus
	}

	rule.Reject()
	if err := s.db.UpdateStatus(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd server && go test ./services/library -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/services/library/service.go server/services/library/service_test.go
git commit -m "feat(service): add library service for centralized rule management"
```

---

## Phase 4: API Layer

### Task 7: Create Library Handler

**Files:**
- Create: `server/entrypoints/api/handlers/library.go`
- Create: `server/entrypoints/api/handlers/library_test.go`

**Step 1: Write the handler**

```go
// server/entrypoints/api/handlers/library.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/library"
)

type LibraryService interface {
	Create(ctx context.Context, req library.CreateRequest) (domain.Rule, error)
	GetByID(ctx context.Context, id string) (domain.Rule, error)
	List(ctx context.Context) ([]domain.Rule, error)
	Update(ctx context.Context, rule domain.Rule) error
	Delete(ctx context.Context, id string) error
	Submit(ctx context.Context, id string) (domain.Rule, error)
	Approve(ctx context.Context, id, approvedBy string) (domain.Rule, error)
	Reject(ctx context.Context, id string) (domain.Rule, error)
}

type LibraryHandler struct {
	service LibraryService
}

func NewLibraryHandler(service LibraryService) *LibraryHandler {
	return &LibraryHandler{service: service}
}

type CreateLibraryRuleRequest struct {
	Name           string           `json:"name"`
	Content        string           `json:"content"`
	Description    string           `json:"description,omitempty"`
	TargetLayer    string           `json:"target_layer"`
	CategoryID     string           `json:"category_id,omitempty"`
	PriorityWeight int              `json:"priority_weight"`
	Overridable    bool             `json:"overridable"`
	Tags           []string         `json:"tags,omitempty"`
	Triggers       []TriggerRequest `json:"triggers,omitempty"`
}

func (h *LibraryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateLibraryRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())

	var triggers []domain.Trigger
	for _, t := range req.Triggers {
		triggers = append(triggers, domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		})
	}

	rule, err := h.service.Create(r.Context(), library.CreateRequest{
		Name:           req.Name,
		Content:        req.Content,
		Description:    req.Description,
		TargetLayer:    domain.TargetLayer(req.TargetLayer),
		CategoryID:     req.CategoryID,
		PriorityWeight: req.PriorityWeight,
		Overridable:    req.Overridable,
		Tags:           req.Tags,
		Triggers:       triggers,
		CreatedBy:      userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rule, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []RuleResponse
	for _, rule := range rules {
		response = append(response, ruleToResponse(rule))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LibraryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req CreateLibraryRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rule.Name = req.Name
	rule.Content = req.Content
	if req.Description != "" {
		rule.Description = &req.Description
	}
	rule.TargetLayer = domain.TargetLayer(req.TargetLayer)
	if req.CategoryID != "" {
		rule.CategoryID = &req.CategoryID
	}
	rule.PriorityWeight = req.PriorityWeight
	rule.Overridable = req.Overridable
	rule.Tags = req.Tags

	rule.Triggers = nil
	for _, t := range req.Triggers {
		rule.Triggers = append(rule.Triggers, domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		})
	}

	if err := h.service.Update(r.Context(), rule); err != nil {
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "can only edit draft or rejected rules", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "can only delete draft rules", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) Submit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.service.Submit(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "rule cannot be submitted in current status", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	rule, err := h.service.Approve(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "only pending rules can be approved", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.service.Reject(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "only pending rules can be rejected", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/submit", h.Submit)
	r.Post("/{id}/approve", h.Approve)
	r.Post("/{id}/reject", h.Reject)
}
```

**Step 2: Verify compilation**

Run: `cd server && go build ./entrypoints/api/handlers`
Expected: No errors

**Step 3: Commit**

```bash
git add server/entrypoints/api/handlers/library.go
git commit -m "feat(api): add library handler for rule library endpoints"
```

---

### Task 8: Create Attachments Handler

**Files:**
- Create: `server/entrypoints/api/handlers/attachments.go`

**Step 1: Write the handler**

```go
// server/entrypoints/api/handlers/attachments.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/attachments"
)

type AttachmentService interface {
	RequestAttachment(ctx context.Context, req attachments.AttachRequest) (domain.RuleAttachment, error)
	GetByID(ctx context.Context, id string) (domain.RuleAttachment, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error)
	ApproveAttachment(ctx context.Context, id, approvedBy string) (domain.RuleAttachment, error)
	RejectAttachment(ctx context.Context, id string) (domain.RuleAttachment, error)
	UpdateEnforcement(ctx context.Context, id string, mode domain.EnforcementMode, timeoutHours int) (domain.RuleAttachment, error)
	Delete(ctx context.Context, id string) error
}

type AttachmentsHandler struct {
	service AttachmentService
}

func NewAttachmentsHandler(service AttachmentService) *AttachmentsHandler {
	return &AttachmentsHandler{service: service}
}

type AttachmentResponse struct {
	ID                    string `json:"id"`
	RuleID                string `json:"ruleId"`
	TeamID                string `json:"teamId"`
	EnforcementMode       string `json:"enforcementMode"`
	TemporaryTimeoutHours int    `json:"temporaryTimeoutHours"`
	Status                string `json:"status"`
	RequestedBy           string `json:"requestedBy"`
	ApprovedBy            string `json:"approvedBy,omitempty"`
	CreatedAt             string `json:"createdAt"`
	ApprovedAt            string `json:"approvedAt,omitempty"`
}

func attachmentToResponse(att domain.RuleAttachment) AttachmentResponse {
	resp := AttachmentResponse{
		ID:                    att.ID,
		RuleID:                att.RuleID,
		TeamID:                att.TeamID,
		EnforcementMode:       string(att.EnforcementMode),
		TemporaryTimeoutHours: att.TemporaryTimeoutHours,
		Status:                string(att.Status),
		RequestedBy:           att.RequestedBy,
		CreatedAt:             att.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if att.ApprovedBy != nil {
		resp.ApprovedBy = *att.ApprovedBy
	}
	if att.ApprovedAt != nil {
		resp.ApprovedAt = att.ApprovedAt.Format("2006-01-02T15:04:05Z")
	}
	return resp
}

type CreateAttachmentRequest struct {
	RuleID          string `json:"rule_id"`
	EnforcementMode string `json:"enforcement_mode"`
	TimeoutHours    int    `json:"temporary_timeout_hours,omitempty"`
}

func (h *AttachmentsHandler) CreateForTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamId")
	userID := middleware.GetUserID(r.Context())

	var req CreateAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mode := domain.EnforcementMode(req.EnforcementMode)
	if !mode.IsValid() {
		http.Error(w, "invalid enforcement_mode", http.StatusBadRequest)
		return
	}

	att, err := h.service.RequestAttachment(r.Context(), attachments.AttachRequest{
		RuleID:          req.RuleID,
		TeamID:          teamID,
		EnforcementMode: mode,
		TimeoutHours:    req.TimeoutHours,
		RequestedBy:     userID,
	})
	if err != nil {
		if errors.Is(err, attachments.ErrRuleNotApproved) {
			http.Error(w, "rule must be approved before attaching", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) ListByTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamId")

	atts, err := h.service.ListByTeam(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []AttachmentResponse
	for _, att := range atts {
		response = append(response, attachmentToResponse(att))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AttachmentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	att, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

type UpdateEnforcementRequest struct {
	EnforcementMode string `json:"enforcement_mode"`
	TimeoutHours    int    `json:"temporary_timeout_hours,omitempty"`
}

func (h *AttachmentsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateEnforcementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mode := domain.EnforcementMode(req.EnforcementMode)
	if !mode.IsValid() {
		http.Error(w, "invalid enforcement_mode", http.StatusBadRequest)
		return
	}

	att, err := h.service.UpdateEnforcement(r.Context(), id, mode, req.TimeoutHours)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AttachmentsHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	att, err := h.service.ApproveAttachment(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	att, err := h.service.RejectAttachment(r.Context(), id)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) RegisterTeamRoutes(r chi.Router) {
	r.Post("/", h.CreateForTeam)
	r.Get("/", h.ListByTeam)
}

func (h *AttachmentsHandler) RegisterAttachmentRoutes(r chi.Router) {
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/approve", h.Approve)
	r.Post("/{id}/reject", h.Reject)
}
```

**Step 2: Verify compilation**

Run: `cd server && go build ./entrypoints/api/handlers`
Expected: No errors

**Step 3: Commit**

```bash
git add server/entrypoints/api/handlers/attachments.go
git commit -m "feat(api): add attachments handler for rule attachment endpoints"
```

---

## Phase 5: Frontend

### Task 9: Add RuleAttachment Domain Type

**Files:**
- Modify: `web/src/domain/rule.ts`

**Step 1: Add the attachment types**

Add to `web/src/domain/rule.ts`:

```typescript
export type AttachmentStatus = 'pending' | 'approved' | 'rejected';

export interface RuleAttachment {
  id: string;
  ruleId: string;
  teamId: string;
  enforcementMode: EnforcementMode;
  temporaryTimeoutHours: number;
  status: AttachmentStatus;
  requestedBy: string;
  approvedBy?: string;
  createdAt: string;
  approvedAt?: string;
}

export interface RuleWithAttachment extends Rule {
  attachment?: RuleAttachment;
}

export function getAttachmentStatusColor(status: AttachmentStatus): string {
  switch (status) {
    case 'pending':
      return 'bg-yellow-100 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-400';
    case 'approved':
      return 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-400';
    case 'rejected':
      return 'bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-400';
    default:
      return 'bg-zinc-100 dark:bg-zinc-700 text-zinc-800 dark:text-zinc-300';
  }
}
```

**Step 2: Commit**

```bash
git add web/src/domain/rule.ts
git commit -m "feat(web): add RuleAttachment domain types"
```

---

### Task 10: Add Library API Functions

**Files:**
- Create: `web/src/lib/api/library.ts`
- Create: `web/src/lib/api/attachments.ts`

**Step 1: Create library API**

```typescript
// web/src/lib/api/library.ts
import { Rule } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchLibraryRules(): Promise<Rule[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch library rules');
  return res.json() || [];
}

export async function fetchLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch library rule');
  return res.json();
}

export interface CreateLibraryRuleRequest {
  name: string;
  content: string;
  description?: string;
  target_layer: string;
  category_id?: string;
  priority_weight?: number;
  overridable?: boolean;
  tags?: string[];
  triggers?: Array<{
    type: string;
    pattern?: string;
    context_types?: string[];
    tags?: string[];
  }>;
}

export async function createLibraryRule(request: CreateLibraryRuleRequest): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) throw new Error('Failed to create library rule');
  return res.json();
}

export async function updateLibraryRule(id: string, request: CreateLibraryRuleRequest): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) throw new Error('Failed to update library rule');
}

export async function deleteLibraryRule(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to delete library rule');
}

export async function submitLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}/submit`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to submit library rule');
  return res.json();
}

export async function approveLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to approve library rule');
  return res.json();
}

export async function rejectLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}/reject`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to reject library rule');
  return res.json();
}
```

**Step 2: Create attachments API**

```typescript
// web/src/lib/api/attachments.ts
import { RuleAttachment, EnforcementMode } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchTeamAttachments(teamId: string): Promise<RuleAttachment[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${teamId}/attachments`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch team attachments');
  return res.json() || [];
}

export interface CreateAttachmentRequest {
  rule_id: string;
  enforcement_mode: EnforcementMode;
  temporary_timeout_hours?: number;
}

export async function createAttachment(teamId: string, request: CreateAttachmentRequest): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${teamId}/attachments`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) throw new Error('Failed to create attachment');
  return res.json();
}

export async function fetchAttachment(id: string): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch attachment');
  return res.json();
}

export async function updateAttachment(id: string, enforcementMode: EnforcementMode, timeoutHours?: number): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      enforcement_mode: enforcementMode,
      temporary_timeout_hours: timeoutHours,
    }),
  });
  if (!res.ok) throw new Error('Failed to update attachment');
  return res.json();
}

export async function deleteAttachment(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to delete attachment');
}

export async function approveAttachment(id: string): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to approve attachment');
  return res.json();
}

export async function rejectAttachment(id: string): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}/reject`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to reject attachment');
  return res.json();
}
```

**Step 3: Commit**

```bash
git add web/src/lib/api/library.ts web/src/lib/api/attachments.ts
git commit -m "feat(web): add library and attachments API functions"
```

---

## Phase 6: Data Migration

### Task 11: Create Data Migration Script

**Files:**
- Create: `server/migrations/000010_migrate_rules_to_library.up.sql`
- Create: `server/migrations/000010_migrate_rules_to_library.down.sql`

**Step 1: Write the data migration**

```sql
-- server/migrations/000010_migrate_rules_to_library.up.sql
-- Migrate existing team rules to library model with attachments

-- Step 1: Create attachments for existing team rules
INSERT INTO rule_attachments (id, rule_id, team_id, enforcement_mode, temporary_timeout_hours, status, requested_by, approved_by, created_at, approved_at)
SELECT
    gen_random_uuid(),
    r.id,
    r.team_id,
    r.enforcement_mode,
    r.temporary_timeout_hours,
    CASE WHEN r.status = 'approved' THEN 'approved' ELSE 'pending' END,
    COALESCE(r.created_by, (SELECT id FROM users LIMIT 1)),
    CASE WHEN r.status = 'approved' THEN r.created_by ELSE NULL END,
    r.created_at,
    CASE WHEN r.status = 'approved' THEN r.approved_at ELSE NULL END
FROM rules r
WHERE r.team_id IS NOT NULL;

-- Step 2: Create attachments for enterprise/forced rules (attach to all teams)
INSERT INTO rule_attachments (id, rule_id, team_id, enforcement_mode, temporary_timeout_hours, status, requested_by, approved_by, created_at, approved_at)
SELECT
    gen_random_uuid(),
    r.id,
    t.id,
    r.enforcement_mode,
    r.temporary_timeout_hours,
    'approved',
    COALESCE(r.created_by, (SELECT id FROM users LIMIT 1)),
    r.created_by,
    r.created_at,
    r.approved_at
FROM rules r
CROSS JOIN teams t
WHERE r.team_id IS NULL AND r.status = 'approved';

-- Step 3: Nullify team_id on all rules (they're now library rules)
-- Note: We don't drop the column yet to allow rollback
UPDATE rules SET team_id = NULL;
```

**Step 2: Write the rollback migration**

```sql
-- server/migrations/000010_migrate_rules_to_library.down.sql
-- Restore team ownership from attachments

-- Step 1: Restore team_id from first attachment for each rule
UPDATE rules r
SET team_id = (
    SELECT ra.team_id
    FROM rule_attachments ra
    WHERE ra.rule_id = r.id
    ORDER BY ra.created_at ASC
    LIMIT 1
);

-- Step 2: Delete all attachments
DELETE FROM rule_attachments;
```

**Step 3: Commit**

```bash
git add server/migrations/000010_migrate_rules_to_library.up.sql server/migrations/000010_migrate_rules_to_library.down.sql
git commit -m "feat(db): add data migration for rule library transition"
```

---

## Phase 7: Integration & Wiring

### Task 12: Wire Up Services and Routes

**Files:**
- Modify: `server/cmd/master/services.go` (or equivalent main wiring file)

**Step 1: Find the main wiring file**

Run: `cat server/cmd/master/services.go | head -100`

**Step 2: Add service initialization**

Add to the services initialization:

```go
// Rule attachment database
ruleAttachmentDB := postgres.NewRuleAttachmentDB(pool)

// Attachments service
attachmentsSvc := attachments.NewService(ruleAttachmentDB, ruleDB, teamDB)

// Library service
librarySvc := library.NewService(ruleDB, attachmentsSvc)

// Handlers
libraryHandler := handlers.NewLibraryHandler(librarySvc)
attachmentsHandler := handlers.NewAttachmentsHandler(attachmentsSvc)
```

**Step 3: Add route registration**

```go
// Library routes
r.Route("/api/v1/library/rules", func(r chi.Router) {
    r.Use(authMiddleware)
    libraryHandler.RegisterRoutes(r)
})

// Team attachment routes
r.Route("/api/v1/teams/{teamId}/attachments", func(r chi.Router) {
    r.Use(authMiddleware)
    attachmentsHandler.RegisterTeamRoutes(r)
})

// Attachment management routes
r.Route("/api/v1/attachments", func(r chi.Router) {
    r.Use(authMiddleware)
    attachmentsHandler.RegisterAttachmentRoutes(r)
})
```

**Step 4: Commit**

```bash
git add server/cmd/master/services.go
git commit -m "feat(server): wire up library and attachments services"
```

---

## Phase 8: Testing & Validation

### Task 13: Run All Tests

**Step 1: Run domain tests**

Run: `cd server && go test ./domain/... -v`
Expected: All tests pass

**Step 2: Run service tests**

Run: `cd server && go test ./services/... -v`
Expected: All tests pass

**Step 3: Run handler tests**

Run: `cd server && go test ./entrypoints/... -v`
Expected: All tests pass

**Step 4: Run full test suite**

Run: `cd server && go test ./... -v`
Expected: All tests pass

**Step 5: Run migrations**

Run: `task migrate-up` (or equivalent)
Expected: Migrations apply successfully

**Step 6: Commit any fixes**

```bash
git add -A
git commit -m "fix: address test failures and migration issues"
```

---

## Summary

This plan implements the Rule Library feature in 13 tasks across 8 phases:

1. **Domain Layer** (Tasks 1-2): New `RuleAttachment` entity and updated `Rule` entity
2. **Database Layer** (Tasks 3-4): Migration for `rule_attachments` table and DB adapter
3. **Service Layer** (Tasks 5-6): `AttachmentsService` and `LibraryService`
4. **API Layer** (Tasks 7-8): `LibraryHandler` and `AttachmentsHandler`
5. **Frontend** (Tasks 9-10): TypeScript types and API functions
6. **Data Migration** (Task 11): Migrate existing rules to library model
7. **Integration** (Task 12): Wire up services and routes
8. **Testing** (Task 13): Run full test suite

Each task follows TDD with explicit test-first steps and frequent commits.
