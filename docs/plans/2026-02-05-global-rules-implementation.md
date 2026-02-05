# Global Rules Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add organization-wide global rules with no team ownership, selective team inheritance, and force flag for mandatory enforcement.

**Architecture:** Global rules have `team_id = NULL` and `force` boolean. Teams have `inherit_global_rules` setting in TeamSettings. Query logic includes forced rules for all teams, and inheritable rules only for teams that opt-in.

**Tech Stack:** Go (domain/handlers/db), PostgreSQL (migrations), TypeScript/React (web UI)

---

## Task 1: Add Force Field to Rule Domain Model

**Files:**
- Modify: `server/domain/rule.go:83-107`
- Test: `server/domain/rule_test.go`

**Step 1: Write failing test for IsGlobal method**

Add to `server/domain/rule_test.go`:

```go
func TestRule_IsGlobal(t *testing.T) {
	tests := []struct {
		name   string
		teamID *string
		want   bool
	}{
		{
			name:   "nil team_id is global",
			teamID: nil,
			want:   true,
		},
		{
			name:   "empty team_id is global",
			teamID: strPtr(""),
			want:   true,
		},
		{
			name:   "non-empty team_id is not global",
			teamID: strPtr("team-123"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := domain.Rule{TeamID: tt.teamID}
			if got := rule.IsGlobal(); got != tt.want {
				t.Errorf("Rule.IsGlobal() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain -run TestRule_IsGlobal -v`
Expected: FAIL - `rule.TeamID` is `string`, not `*string`

**Step 3: Update Rule struct - change TeamID to pointer and add Force field**

In `server/domain/rule.go`, update the Rule struct:

```go
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
	SubmittedAt           *time.Time      `json:"submitted_at,omitempty"`
	ApprovedAt            *time.Time      `json:"approved_at,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}
```

**Step 4: Add IsGlobal method**

Add after the Rule struct in `server/domain/rule.go`:

```go
// IsGlobal returns true if this is a global rule (no team ownership)
func (r *Rule) IsGlobal() bool {
	return r.TeamID == nil || *r.TeamID == ""
}
```

**Step 5: Run test to verify it passes**

Run: `cd server && go test ./domain -run TestRule_IsGlobal -v`
Expected: PASS

**Step 6: Commit**

```bash
git add server/domain/rule.go server/domain/rule_test.go
git commit -m "feat(domain): add Force field and IsGlobal method to Rule"
```

---

## Task 2: Add Force Validation to Rule

**Files:**
- Modify: `server/domain/rule.go`
- Test: `server/domain/rule_test.go`

**Step 1: Write failing test for force validation**

Add to `server/domain/rule_test.go`:

```go
func TestRule_ValidateForce(t *testing.T) {
	teamID := "team-123"
	tests := []struct {
		name    string
		rule    domain.Rule
		wantErr bool
	}{
		{
			name: "global rule with force=true is valid",
			rule: domain.Rule{
				Name:        "Forced Global Rule",
				Content:     "content",
				TargetLayer: domain.TargetLayerEnterprise,
				TeamID:      nil,
				Force:       true,
			},
			wantErr: false,
		},
		{
			name: "global rule with force=false is valid",
			rule: domain.Rule{
				Name:        "Inheritable Global Rule",
				Content:     "content",
				TargetLayer: domain.TargetLayerEnterprise,
				TeamID:      nil,
				Force:       false,
			},
			wantErr: false,
		},
		{
			name: "team rule with force=true is invalid",
			rule: domain.Rule{
				Name:        "Team Rule",
				Content:     "content",
				TargetLayer: domain.TargetLayerProject,
				TeamID:      &teamID,
				Force:       true,
			},
			wantErr: true,
		},
		{
			name: "global rule must be enterprise layer",
			rule: domain.Rule{
				Name:        "Global Project Rule",
				Content:     "content",
				TargetLayer: domain.TargetLayerProject,
				TeamID:      nil,
				Force:       false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain -run TestRule_ValidateForce -v`
Expected: FAIL - validation doesn't check force constraints yet

**Step 3: Update Validate method**

In `server/domain/rule.go`, update the `Validate` method:

```go
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
		if r.TargetLayer != TargetLayerEnterprise {
			return errors.New("global rules must have enterprise target layer")
		}
	} else {
		// Team rule constraints
		if r.Force {
			return errors.New("force flag is only valid for global rules")
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain -run TestRule_ValidateForce -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/rule.go server/domain/rule_test.go
git commit -m "feat(domain): add force validation to Rule.Validate"
```

---

## Task 3: Add InheritGlobalRules to TeamSettings

**Files:**
- Modify: `server/domain/team.go`
- Test: `server/domain/team_test.go`

**Step 1: Write failing test for default InheritGlobalRules**

Add to `server/domain/team_test.go`:

```go
func TestNewTeam_DefaultsInheritGlobalRules(t *testing.T) {
	team := domain.NewTeam("Engineering")

	if !team.Settings.InheritGlobalRules {
		t.Error("expected InheritGlobalRules to default to true")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain -run TestNewTeam_DefaultsInheritGlobalRules -v`
Expected: FAIL - `InheritGlobalRules` field doesn't exist

**Step 3: Update TeamSettings struct**

In `server/domain/team.go`:

```go
type TeamSettings struct {
	DriftThresholdMinutes int  `json:"drift_threshold_minutes"`
	InheritGlobalRules    bool `json:"inherit_global_rules"`
}
```

**Step 4: Update NewTeam function**

In `server/domain/team.go`:

```go
func NewTeam(name string) Team {
	return Team{
		ID:   uuid.New().String(),
		Name: name,
		Settings: TeamSettings{
			DriftThresholdMinutes: 60,
			InheritGlobalRules:    true,
		},
		CreatedAt: time.Now(),
	}
}
```

**Step 5: Run test to verify it passes**

Run: `cd server && go test ./domain -run TestNewTeam_DefaultsInheritGlobalRules -v`
Expected: PASS

**Step 6: Commit**

```bash
git add server/domain/team.go server/domain/team_test.go
git commit -m "feat(domain): add InheritGlobalRules to TeamSettings"
```

---

## Task 4: Update NewRule Function for Optional TeamID

**Files:**
- Modify: `server/domain/rule.go`
- Test: `server/domain/rule_test.go`

**Step 1: Write test for NewGlobalRule constructor**

Add to `server/domain/rule_test.go`:

```go
func TestNewGlobalRule(t *testing.T) {
	rule := domain.NewGlobalRule("Security Policy", "Never hardcode secrets", true)

	if !rule.IsGlobal() {
		t.Error("expected global rule to have nil TeamID")
	}
	if rule.TargetLayer != domain.TargetLayerEnterprise {
		t.Errorf("expected enterprise layer, got %s", rule.TargetLayer)
	}
	if !rule.Force {
		t.Error("expected force to be true")
	}
}

func TestNewRule_WithTeamID(t *testing.T) {
	triggers := []domain.Trigger{}
	rule := domain.NewRule("Team Rule", domain.TargetLayerProject, "content", triggers, "team-123")

	if rule.IsGlobal() {
		t.Error("expected team rule to have TeamID set")
	}
	if rule.TeamID == nil || *rule.TeamID != "team-123" {
		t.Errorf("expected TeamID 'team-123', got %v", rule.TeamID)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd server && go test ./domain -run "TestNewGlobalRule|TestNewRule_WithTeamID" -v`
Expected: FAIL - NewGlobalRule doesn't exist, NewRule uses string not *string

**Step 3: Update NewRule and add NewGlobalRule**

In `server/domain/rule.go`:

```go
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
		TargetLayer:           TargetLayerEnterprise,
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
```

**Step 4: Run tests to verify they pass**

Run: `cd server && go test ./domain -run "TestNewGlobalRule|TestNewRule_WithTeamID" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/rule.go server/domain/rule_test.go
git commit -m "feat(domain): add NewGlobalRule constructor"
```

---

## Task 5: Fix Compilation Errors from TeamID Change

**Files:**
- Modify: `server/adapters/postgres/rule_db.go`
- Modify: `server/entrypoints/api/handlers/rules.go`
- Modify: `server/services/rules/repository.go`

**Step 1: Run build to identify errors**

Run: `cd server && go build ./...`
Expected: Multiple compilation errors due to TeamID type change

**Step 2: Fix rule_db.go - scan into pointer**

Update `scanRules` in `server/adapters/postgres/rule_db.go` to use a nullable scan variable and convert appropriately. The scan should handle NULL team_id values.

**Step 3: Fix handlers/rules.go - handle optional team_id in response**

Update `ruleToResponse` to handle nullable TeamID:

```go
func ruleToResponse(rule domain.Rule) RuleResponse {
	resp := RuleResponse{
		ID:             rule.ID,
		Name:           rule.Name,
		Content:        rule.Content,
		TargetLayer:    string(rule.TargetLayer),
		PriorityWeight: rule.PriorityWeight,
		Status:         string(rule.Status),
		CreatedBy:      rule.CreatedBy,
		Force:          rule.Force,
		CreatedAt:      rule.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      rule.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if rule.TeamID != nil {
		resp.TeamID = *rule.TeamID
	}
	// ... rest of method
}
```

**Step 4: Run build to verify no errors**

Run: `cd server && go build ./...`
Expected: SUCCESS

**Step 5: Run all tests**

Run: `cd server && go test ./... -v`
Expected: PASS (some tests may need fixing due to TeamID change)

**Step 6: Commit**

```bash
git add server/adapters/postgres/rule_db.go server/entrypoints/api/handlers/rules.go
git commit -m "fix: update code for nullable TeamID"
```

---

## Task 6: Create Database Migration

**Files:**
- Create: `server/migrations/000023_add_global_rules.up.sql`
- Create: `server/migrations/000023_add_global_rules.down.sql`

**Step 1: Create up migration**

Create `server/migrations/000023_add_global_rules.up.sql`:

```sql
-- Make team_id nullable for global rules
ALTER TABLE rules ALTER COLUMN team_id DROP NOT NULL;

-- Add force column for global rules
ALTER TABLE rules ADD COLUMN force BOOLEAN NOT NULL DEFAULT false;

-- Constraint: force only valid for global rules (team_id IS NULL)
ALTER TABLE rules ADD CONSTRAINT rules_force_global_only
    CHECK (force = false OR team_id IS NULL);

-- Constraint: global rules must be enterprise layer
ALTER TABLE rules ADD CONSTRAINT rules_global_enterprise_only
    CHECK (team_id IS NOT NULL OR target_layer = 'enterprise');

-- Index for efficient global rule queries
CREATE INDEX idx_rules_global ON rules(force) WHERE team_id IS NULL;

-- Index for force flag queries
CREATE INDEX idx_rules_force ON rules(force) WHERE force = true;
```

**Step 2: Create down migration**

Create `server/migrations/000023_add_global_rules.down.sql`:

```sql
-- Remove indexes
DROP INDEX IF EXISTS idx_rules_force;
DROP INDEX IF EXISTS idx_rules_global;

-- Remove constraints
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_global_enterprise_only;
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_force_global_only;

-- Remove force column
ALTER TABLE rules DROP COLUMN IF EXISTS force;

-- Restore team_id NOT NULL (will fail if global rules exist)
-- Must delete global rules first: DELETE FROM rules WHERE team_id IS NULL;
ALTER TABLE rules ALTER COLUMN team_id SET NOT NULL;
```

**Step 3: Commit**

```bash
git add server/migrations/000023_add_global_rules.up.sql server/migrations/000023_add_global_rules.down.sql
git commit -m "feat(db): add migration for global rules"
```

---

## Task 7: Update RuleDB for Global Rules

**Files:**
- Modify: `server/adapters/postgres/rule_db.go`
- Test: `server/integration/repository_test.go` (if exists)

**Step 1: Add ListGlobalRules method**

Add to `server/adapters/postgres/rule_db.go`:

```go
// ListGlobalRules retrieves all global rules (team_id IS NULL)
func (db *RuleDB) ListGlobalRules(ctx context.Context) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE team_id IS NULL
		ORDER BY force DESC, priority_weight DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}
```

**Step 2: Update GetRulesForMerge for global rules with team inheritance**

Update in `server/adapters/postgres/rule_db.go`:

```go
// GetRulesForMerge returns all approved rules for a given target layer, filtered by targeting
// For global rules, respects team's inherit_global_rules setting
func (db *RuleDB) GetRulesForMerge(ctx context.Context, targetLayer domain.TargetLayer, userID string, teamIDs []string, teamInheritsGlobal bool) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE status = 'approved'
		  AND (
			-- Global rules
			(team_id IS NULL AND target_layer = $1 AND (force = true OR $4 = true))
			OR
			-- Team rules (existing logic)
			(team_id IS NOT NULL AND target_layer = $1 AND (
				(target_teams = '{}' AND target_users = '{}')
				OR $2 = ANY(target_users)
				OR target_teams && $3::uuid[]
			))
		  )
		ORDER BY
			CASE WHEN team_id IS NULL THEN 0 ELSE 1 END,
			force DESC,
			priority_weight DESC,
			name
	`, targetLayer, userID, teamIDs, teamInheritsGlobal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}
```

**Step 3: Update scanRules to handle force column**

Ensure the `scanRules` method scans the `force` column (add it between `team_id` and `status` in the scan order).

**Step 4: Commit**

```bash
git add server/adapters/postgres/rule_db.go
git commit -m "feat(db): add ListGlobalRules and update GetRulesForMerge"
```

---

## Task 8: Add Global Rules API Endpoint

**Files:**
- Modify: `server/entrypoints/api/handlers/rules.go`
- Test: `server/tests/unit/handlers/rules_test.go`

**Step 1: Update RuleService interface**

Add to interface in `server/entrypoints/api/handlers/rules.go`:

```go
type RuleService interface {
	// ... existing methods
	ListGlobal(ctx context.Context) ([]domain.Rule, error)
	CreateGlobal(ctx context.Context, req CreateGlobalRuleRequest) (domain.Rule, error)
}
```

**Step 2: Add CreateGlobalRuleRequest type**

Add to `server/entrypoints/api/handlers/rules.go`:

```go
type CreateGlobalRuleRequest struct {
	Name        string           `json:"name"`
	Content     string           `json:"content"`
	Description *string          `json:"description,omitempty"`
	Force       bool             `json:"force"`
	Triggers    []TriggerRequest `json:"triggers,omitempty"`
}
```

**Step 3: Add RuleResponse Force field**

Update `RuleResponse` struct:

```go
type RuleResponse struct {
	// ... existing fields
	TeamID  string `json:"team_id,omitempty"`
	Force   bool   `json:"force"`
	// ... rest of fields
}
```

**Step 4: Add ListGlobal handler**

Add to `server/entrypoints/api/handlers/rules.go`:

```go
func (h *RulesHandler) ListGlobal(w http.ResponseWriter, r *http.Request) {
	rules, err := h.service.ListGlobal(r.Context())
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
```

**Step 5: Add CreateGlobal handler**

Add to `server/entrypoints/api/handlers/rules.go`:

```go
func (h *RulesHandler) CreateGlobal(w http.ResponseWriter, r *http.Request) {
	var req CreateGlobalRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rule, err := h.service.CreateGlobal(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.publisher != nil {
		go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleCreated, rule.ID, "")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}
```

**Step 6: Register routes**

Update `RegisterRoutes` in `server/entrypoints/api/handlers/rules.go`:

```go
func (h *RulesHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.ListByTeam)
	r.Get("/global", h.ListGlobal)      // New
	r.Post("/global", h.CreateGlobal)   // New
	r.Get("/merged", h.GetMerged)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Patch("/{id}", h.UpdateEnforcement)
	r.Delete("/{id}", h.Delete)
}
```

**Step 7: Commit**

```bash
git add server/entrypoints/api/handlers/rules.go
git commit -m "feat(api): add global rules endpoints"
```

---

## Task 9: Add Team Settings Update Endpoint

**Files:**
- Modify: `server/entrypoints/api/handlers/teams.go`
- Test: `server/tests/unit/handlers/teams_test.go`

**Step 1: Add UpdateSettings handler**

Add to `server/entrypoints/api/handlers/teams.go`:

```go
type UpdateTeamSettingsRequest struct {
	DriftThresholdMinutes *int  `json:"drift_threshold_minutes,omitempty"`
	InheritGlobalRules    *bool `json:"inherit_global_rules,omitempty"`
}

func (h *TeamsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	team, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req UpdateTeamSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DriftThresholdMinutes != nil {
		team.Settings.DriftThresholdMinutes = *req.DriftThresholdMinutes
	}
	if req.InheritGlobalRules != nil {
		team.Settings.InheritGlobalRules = *req.InheritGlobalRules
	}

	if err := h.service.Update(r.Context(), team); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}
```

**Step 2: Register route**

Update `RegisterRoutes` in `server/entrypoints/api/handlers/teams.go`:

```go
func (h *TeamsHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}/settings", h.UpdateSettings)  // New
	r.Delete("/{id}", h.Delete)
}
```

**Step 3: Commit**

```bash
git add server/entrypoints/api/handlers/teams.go
git commit -m "feat(api): add team settings update endpoint"
```

---

## Task 10: Update Web Domain Types

**Files:**
- Modify: `web/src/domain/rule.ts`

**Step 1: Update Rule interface**

In `web/src/domain/rule.ts`:

```typescript
export interface Rule {
  id: string;
  name: string;
  content: string;
  description?: string;
  targetLayer: TargetLayer;
  categoryId?: string;
  priorityWeight: number;
  overridable: boolean;
  effectiveStart?: string;
  effectiveEnd?: string;
  targetTeams?: string[];
  targetUsers?: string[];
  tags?: string[];
  triggers: Trigger[];
  teamId?: string;           // Made optional for global rules
  force: boolean;            // New field
  status: RuleStatus;
  enforcementMode: EnforcementMode;
  temporaryTimeoutHours: number;
  createdBy?: string;
  submittedAt?: string;
  approvedAt?: string;
  createdAt: string;
  updatedAt: string;
}
```

**Step 2: Add helper functions**

Add to `web/src/domain/rule.ts`:

```typescript
export function isGlobalRule(rule: Rule): boolean {
  return !rule.teamId || rule.teamId === '';
}

export function getEnforcementLabel(rule: Rule): string {
  if (!isGlobalRule(rule)) {
    return 'Team';
  }
  return rule.force ? 'Forced' : 'Inheritable';
}

export function getEnforcementColor(rule: Rule): string {
  if (!isGlobalRule(rule)) {
    return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300';
  }
  if (rule.force) {
    return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300';
  }
  return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300';
}
```

**Step 3: Add TeamSettings interface**

Add to `web/src/domain/rule.ts` or create new file `web/src/domain/team.ts`:

```typescript
export interface TeamSettings {
  drift_threshold_minutes: number;
  inherit_global_rules: boolean;
}

export interface Team {
  id: string;
  name: string;
  settings: TeamSettings;
  createdAt: string;
}
```

**Step 4: Commit**

```bash
git add web/src/domain/rule.ts
git commit -m "feat(web): add force field and global rule helpers"
```

---

## Task 11: Update Web API Client

**Files:**
- Modify: `web/src/lib/api.ts`

**Step 1: Add global rules API functions**

Add to `web/src/lib/api.ts`:

```typescript
export async function fetchGlobalRules(): Promise<Rule[]> {
  const res = await fetch(`${API_BASE}/rules/global`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch global rules');
  return res.json();
}

export async function createGlobalRule(data: {
  name: string;
  content: string;
  description?: string;
  force: boolean;
}): Promise<Rule> {
  const res = await fetch(`${API_BASE}/rules/global`, {
    method: 'POST',
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to create global rule');
  return res.json();
}

export async function updateTeamSettings(
  teamId: string,
  settings: { inherit_global_rules?: boolean; drift_threshold_minutes?: number }
): Promise<Team> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/settings`, {
    method: 'PATCH',
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(settings),
  });
  if (!res.ok) throw new Error('Failed to update team settings');
  return res.json();
}
```

**Step 2: Commit**

```bash
git add web/src/lib/api.ts
git commit -m "feat(web): add global rules API functions"
```

---

## Task 12: Update RuleList Component for Global Tab

**Files:**
- Modify: `web/src/components/RuleList.tsx`

**Step 1: Add global rules tab**

Update `web/src/components/RuleList.tsx` to include:

1. A tab switcher between "Team Rules" and "Global Rules"
2. Fetch global rules when Global tab is selected
3. Display enforcement badge (Forced/Inheritable) for global rules
4. Hide create button for global rules tab (admins use separate admin page)

**Step 2: Add enforcement badge rendering**

```tsx
import { isGlobalRule, getEnforcementLabel, getEnforcementColor } from '@/domain/rule';

// In the rule card/row:
{isGlobalRule(rule) && (
  <span className={`px-2 py-0.5 rounded text-xs font-medium ${getEnforcementColor(rule)}`}>
    {getEnforcementLabel(rule)}
  </span>
)}
```

**Step 3: Commit**

```bash
git add web/src/components/RuleList.tsx
git commit -m "feat(web): add global rules tab to RuleList"
```

---

## Task 13: Update RuleEditor for Global Rules

**Files:**
- Modify: `web/src/components/RuleEditor/index.tsx` (or wherever the form is)

**Step 1: Add global scope option**

When creating a rule, add a "Scope" selector:
- Team (default) - requires team_id selection
- Global (admin only) - shows force checkbox

**Step 2: Add force checkbox**

When scope is Global:
```tsx
<label className="flex items-center gap-2">
  <input
    type="checkbox"
    checked={force}
    onChange={(e) => setForce(e.target.checked)}
  />
  <span>Force on all teams</span>
  <span className="text-xs text-zinc-500">
    (Forced rules apply even to teams that opted out)
  </span>
</label>
```

**Step 3: Commit**

```bash
git add web/src/components/RuleEditor/
git commit -m "feat(web): add global scope option to RuleEditor"
```

---

## Task 14: Add Team Settings Page

**Files:**
- Create: `web/src/app/admin/teams/[id]/settings/page.tsx`

**Step 1: Create settings page**

Create `web/src/app/admin/teams/[id]/settings/page.tsx`:

```tsx
'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { fetchTeam, updateTeamSettings, fetchGlobalRules } from '@/lib/api';
import { Team, Rule } from '@/domain/rule';

export default function TeamSettingsPage() {
  const params = useParams();
  const teamId = params.id as string;
  const [team, setTeam] = useState<Team | null>(null);
  const [forcedRulesCount, setForcedRulesCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      const [teamData, globalRules] = await Promise.all([
        fetchTeam(teamId),
        fetchGlobalRules(),
      ]);
      setTeam(teamData);
      setForcedRulesCount(globalRules.filter(r => r.force).length);
      setLoading(false);
    }
    load();
  }, [teamId]);

  const handleToggleInherit = async () => {
    if (!team) return;
    const updated = await updateTeamSettings(teamId, {
      inherit_global_rules: !team.settings.inherit_global_rules,
    });
    setTeam(updated);
  };

  if (loading || !team) return <div>Loading...</div>;

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">{team.name} Settings</h1>

      <div className="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow">
        <h2 className="text-lg font-semibold mb-4">Global Rules</h2>

        <label className="flex items-center gap-3">
          <input
            type="checkbox"
            checked={team.settings.inherit_global_rules}
            onChange={handleToggleInherit}
            className="w-5 h-5"
          />
          <div>
            <span className="font-medium">Inherit Global Rules</span>
            <p className="text-sm text-zinc-500">
              When disabled, this team will only receive forced global rules
            </p>
          </div>
        </label>

        <p className="mt-4 text-sm text-zinc-600 dark:text-zinc-400">
          {forcedRulesCount} forced rule{forcedRulesCount !== 1 ? 's' : ''} will always apply
        </p>
      </div>
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add web/src/app/admin/teams/
git commit -m "feat(web): add team settings page with inheritance toggle"
```

---

## Task 15: Integration Testing

**Files:**
- Modify: `server/integration/api_test.go`

**Step 1: Add integration test for global rules**

Add to `server/integration/api_test.go`:

```go
func TestGlobalRulesAPI(t *testing.T) {
	// Test creating a global rule
	// Test listing global rules
	// Test that forced rules apply to all teams
	// Test that inheritable rules respect team settings
}
```

**Step 2: Run integration tests**

Run: `cd server && go test ./integration -v -run TestGlobalRules`

**Step 3: Commit**

```bash
git add server/integration/
git commit -m "test: add integration tests for global rules"
```

---

## Task 16: Run Full Test Suite and Fix Issues

**Step 1: Run server tests**

Run: `cd server && go test ./... -v`

**Step 2: Fix any failing tests**

Update tests that assume TeamID is always present.

**Step 3: Run web tests**

Run: `cd web && npm test`

**Step 4: Fix any failing tests**

Update tests for new Rule interface.

**Step 5: Commit fixes**

```bash
git add -A
git commit -m "fix: update tests for global rules changes"
```

---

## Task 17: Final Verification

**Step 1: Start the test stack**

Run: `task test:infra:up`

**Step 2: Run migrations**

The migrations should run automatically on startup.

**Step 3: Manual verification**

1. Login as admin
2. Create a global rule with force=true
3. Create a global rule with force=false
4. Create a team and set inherit_global_rules=false
5. Verify forced rule still applies
6. Verify inheritable rule does not apply

**Step 4: Commit any final fixes**

```bash
git add -A
git commit -m "fix: address issues found in manual testing"
```
