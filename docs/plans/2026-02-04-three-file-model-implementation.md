# Three-File Model Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform Claudeception from managing arbitrary CLAUDE.md files to managing exactly three fixed-location files (Enterprise, User, Project) with rule merging by category.

**Architecture:** Add a categories table, extend the rules table with new fields (description, overridable, effective dates, target_teams, target_users, tags), implement a merge renderer in the agent that builds managed sections from cached rules, and update the WebUI with new form fields and category management.

**Tech Stack:** Go (server + agent), PostgreSQL, SQLite (agent cache), Next.js/React (web), Chi router, fsnotify, gorilla/websocket

---

## Phase 1: Database Schema Changes

### Task 1: Create Categories Table Migration

**Files:**
- Create: `server/migrations/000021_create_categories.up.sql`
- Create: `server/migrations/000021_create_categories.down.sql`

**Step 1: Write the up migration**

```sql
-- 000021_create_categories.up.sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    org_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Unique constraint: name must be unique within org (or globally for system categories)
CREATE UNIQUE INDEX idx_categories_name_org ON categories(name, COALESCE(org_id, '00000000-0000-0000-0000-000000000000'));

-- Insert system default categories
INSERT INTO categories (id, name, is_system, org_id, display_order) VALUES
    (gen_random_uuid(), 'Security', TRUE, NULL, 1),
    (gen_random_uuid(), 'Coding Standards', TRUE, NULL, 2),
    (gen_random_uuid(), 'Testing', TRUE, NULL, 3),
    (gen_random_uuid(), 'Documentation', TRUE, NULL, 4);
```

**Step 2: Write the down migration**

```sql
-- 000021_create_categories.down.sql
DROP TABLE IF EXISTS categories;
```

**Step 3: Commit**

```bash
git add server/migrations/000021_create_categories.up.sql server/migrations/000021_create_categories.down.sql
git commit -m "feat(db): add categories table with system defaults"
```

---

### Task 2: Extend Rules Table Migration

**Files:**
- Create: `server/migrations/000022_extend_rules_three_file.up.sql`
- Create: `server/migrations/000022_extend_rules_three_file.down.sql`

**Step 1: Write the up migration**

```sql
-- 000022_extend_rules_three_file.up.sql

-- Add new columns to rules table
ALTER TABLE rules ADD COLUMN description TEXT;
ALTER TABLE rules ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE SET NULL;
ALTER TABLE rules ADD COLUMN overridable BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE rules ADD COLUMN effective_start TIMESTAMP WITH TIME ZONE;
ALTER TABLE rules ADD COLUMN effective_end TIMESTAMP WITH TIME ZONE;
ALTER TABLE rules ADD COLUMN target_teams UUID[] DEFAULT '{}';
ALTER TABLE rules ADD COLUMN target_users UUID[] DEFAULT '{}';
ALTER TABLE rules ADD COLUMN tags TEXT[] DEFAULT '{}';

-- Update target_layer enum values: rename 'global' to 'user', remove 'local'
-- First update existing values
UPDATE rules SET target_layer = 'user' WHERE target_layer = 'global';
UPDATE rules SET target_layer = 'project' WHERE target_layer = 'local';

-- Add check constraint for new enum values
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_target_layer_check;
ALTER TABLE rules ADD CONSTRAINT rules_target_layer_check
    CHECK (target_layer IN ('enterprise', 'user', 'project'));

-- Create indexes for new columns
CREATE INDEX idx_rules_category_id ON rules(category_id);
CREATE INDEX idx_rules_effective_dates ON rules(effective_start, effective_end);
CREATE INDEX idx_rules_target_teams ON rules USING GIN(target_teams);
CREATE INDEX idx_rules_target_users ON rules USING GIN(target_users);
CREATE INDEX idx_rules_tags ON rules USING GIN(tags);

-- Assign existing rules to a default category (Coding Standards)
UPDATE rules SET category_id = (SELECT id FROM categories WHERE name = 'Coding Standards' AND is_system = TRUE LIMIT 1)
WHERE category_id IS NULL;
```

**Step 2: Write the down migration**

```sql
-- 000022_extend_rules_three_file.down.sql

-- Remove indexes
DROP INDEX IF EXISTS idx_rules_category_id;
DROP INDEX IF EXISTS idx_rules_effective_dates;
DROP INDEX IF EXISTS idx_rules_target_teams;
DROP INDEX IF EXISTS idx_rules_target_users;
DROP INDEX IF EXISTS idx_rules_tags;

-- Remove constraint
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_target_layer_check;

-- Restore original target_layer values
UPDATE rules SET target_layer = 'global' WHERE target_layer = 'user';
UPDATE rules SET target_layer = 'local' WHERE target_layer = 'project';

-- Remove new columns
ALTER TABLE rules DROP COLUMN IF EXISTS description;
ALTER TABLE rules DROP COLUMN IF EXISTS category_id;
ALTER TABLE rules DROP COLUMN IF EXISTS overridable;
ALTER TABLE rules DROP COLUMN IF EXISTS effective_start;
ALTER TABLE rules DROP COLUMN IF EXISTS effective_end;
ALTER TABLE rules DROP COLUMN IF EXISTS target_teams;
ALTER TABLE rules DROP COLUMN IF EXISTS target_users;
ALTER TABLE rules DROP COLUMN IF EXISTS tags;
```

**Step 3: Commit**

```bash
git add server/migrations/000022_extend_rules_three_file.up.sql server/migrations/000022_extend_rules_three_file.down.sql
git commit -m "feat(db): extend rules table for three-file model"
```

---

### Task 3: Run Migrations and Verify

**Step 1: Run migrations**

Run: `cd server && go run cmd/migrate/main.go up`
Expected: Migrations applied successfully

**Step 2: Verify schema**

Run: `psql $DATABASE_URL -c "\d rules" && psql $DATABASE_URL -c "\d categories" && psql $DATABASE_URL -c "SELECT * FROM categories"`
Expected: Both tables exist with correct columns, 4 system categories inserted

**Step 3: Commit any generated files if applicable**

---

## Phase 2: Server Domain Models

### Task 4: Create Category Domain Model

**Files:**
- Create: `server/domain/category.go`
- Test: `server/domain/category_test.go`

**Step 1: Write the failing test**

```go
// server/domain/category_test.go
package domain

import (
    "testing"
)

func TestCategory_Validate(t *testing.T) {
    tests := []struct {
        name     string
        category Category
        wantErr  bool
    }{
        {
            name: "valid system category",
            category: Category{
                Name:     "Security",
                IsSystem: true,
            },
            wantErr: false,
        },
        {
            name: "valid org category",
            category: Category{
                Name:  "Frontend Patterns",
                OrgID: stringPtr("org-123"),
            },
            wantErr: false,
        },
        {
            name: "empty name",
            category: Category{
                Name: "",
            },
            wantErr: true,
        },
        {
            name: "name too long",
            category: Category{
                Name: string(make([]byte, 101)),
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.category.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Category.Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func stringPtr(s string) *string {
    return &s
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain -run TestCategory_Validate -v`
Expected: FAIL - Category type not defined

**Step 3: Write minimal implementation**

```go
// server/domain/category.go
package domain

import (
    "errors"
    "time"
)

type Category struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    IsSystem     bool      `json:"is_system"`
    OrgID        *string   `json:"org_id,omitempty"`
    DisplayOrder int       `json:"display_order"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

func (c *Category) Validate() error {
    if c.Name == "" {
        return errors.New("category name is required")
    }
    if len(c.Name) > 100 {
        return errors.New("category name must be 100 characters or less")
    }
    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain -run TestCategory_Validate -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/category.go server/domain/category_test.go
git commit -m "feat(domain): add Category model with validation"
```

---

### Task 5: Update Rule Domain Model

**Files:**
- Modify: `server/domain/rule.go`
- Modify: `server/domain/rule_test.go`

**Step 1: Write failing tests for new fields**

Add to `server/domain/rule_test.go`:

```go
func TestRule_ValidateOverrideConflict(t *testing.T) {
    tests := []struct {
        name        string
        rule        Rule
        higherRules []Rule
        wantErr     bool
    }{
        {
            name: "no conflict - higher rule is overridable",
            rule: Rule{
                Name:        "Project security rule",
                TargetLayer: TargetLayerProject,
                CategoryID:  stringPtr("cat-1"),
            },
            higherRules: []Rule{
                {
                    Name:        "Enterprise security rule",
                    TargetLayer: TargetLayerEnterprise,
                    CategoryID:  stringPtr("cat-1"),
                    Overridable: true,
                },
            },
            wantErr: false,
        },
        {
            name: "conflict - higher rule not overridable, same category",
            rule: Rule{
                Name:        "Project security rule",
                TargetLayer: TargetLayerProject,
                CategoryID:  stringPtr("cat-1"),
            },
            higherRules: []Rule{
                {
                    Name:        "Enterprise security rule",
                    TargetLayer: TargetLayerEnterprise,
                    CategoryID:  stringPtr("cat-1"),
                    Overridable: false,
                },
            },
            wantErr: true,
        },
        {
            name: "no conflict - different categories",
            rule: Rule{
                Name:        "Project testing rule",
                TargetLayer: TargetLayerProject,
                CategoryID:  stringPtr("cat-2"),
            },
            higherRules: []Rule{
                {
                    Name:        "Enterprise security rule",
                    TargetLayer: TargetLayerEnterprise,
                    CategoryID:  stringPtr("cat-1"),
                    Overridable: false,
                },
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.rule.ValidateOverrideConflict(tt.higherRules)
            if (err != nil) != tt.wantErr {
                t.Errorf("Rule.ValidateOverrideConflict() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestRule_IsEffective(t *testing.T) {
    now := time.Now()
    past := now.Add(-24 * time.Hour)
    future := now.Add(24 * time.Hour)

    tests := []struct {
        name string
        rule Rule
        want bool
    }{
        {
            name: "no dates - always effective",
            rule: Rule{Name: "test"},
            want: true,
        },
        {
            name: "start in past, no end - effective",
            rule: Rule{
                Name:           "test",
                EffectiveStart: &past,
            },
            want: true,
        },
        {
            name: "start in future - not effective",
            rule: Rule{
                Name:           "test",
                EffectiveStart: &future,
            },
            want: false,
        },
        {
            name: "end in past - not effective",
            rule: Rule{
                Name:         "test",
                EffectiveEnd: &past,
            },
            want: false,
        },
        {
            name: "start in past, end in future - effective",
            rule: Rule{
                Name:           "test",
                EffectiveStart: &past,
                EffectiveEnd:   &future,
            },
            want: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.rule.IsEffective(); got != tt.want {
                t.Errorf("Rule.IsEffective() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `cd server && go test ./domain -run "TestRule_ValidateOverrideConflict|TestRule_IsEffective" -v`
Expected: FAIL - new fields/methods not defined

**Step 3: Update Rule struct and add methods**

Update `server/domain/rule.go`:

```go
// Add to imports
import (
    "errors"
    "fmt"
    "time"
)

// Update TargetLayer constants
const (
    TargetLayerEnterprise TargetLayer = "enterprise"
    TargetLayerUser       TargetLayer = "user"
    TargetLayerProject    TargetLayer = "project"
)

// Update Rule struct - add new fields after existing ones
type Rule struct {
    ID                    string          `json:"id"`
    Name                  string          `json:"name"`
    Content               string          `json:"content"`
    Description           string          `json:"description,omitempty"`
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

// Add new methods

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
    case TargetLayerEnterprise:
        return 3
    case TargetLayerUser:
        return 2
    case TargetLayerProject:
        return 1
    default:
        return 0
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `cd server && go test ./domain -run "TestRule_ValidateOverrideConflict|TestRule_IsEffective" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/rule.go server/domain/rule_test.go
git commit -m "feat(domain): extend Rule model with three-file fields"
```

---

## Phase 3: Server Database Layer

### Task 6: Create Category Repository

**Files:**
- Create: `server/adapters/postgres/category_db.go`
- Create: `server/adapters/postgres/category_db_test.go`

**Step 1: Write the failing test**

```go
// server/adapters/postgres/category_db_test.go
package postgres

import (
    "context"
    "testing"

    "claudeception/server/domain"
)

type mockCategoryRow struct {
    id           string
    name         string
    isSystem     bool
    orgID        *string
    displayOrder int
}

func TestCategoryRepository_List(t *testing.T) {
    // This will be an integration test - for now just verify the interface compiles
    var _ CategoryRepository = (*categoryRepository)(nil)
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./adapters/postgres -run TestCategoryRepository -v`
Expected: FAIL - CategoryRepository not defined

**Step 3: Write the implementation**

```go
// server/adapters/postgres/category_db.go
package postgres

import (
    "context"
    "errors"

    "claudeception/server/domain"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var ErrCategoryNotFound = errors.New("category not found")

type CategoryRepository interface {
    Create(ctx context.Context, category domain.Category) (domain.Category, error)
    GetByID(ctx context.Context, id string) (domain.Category, error)
    List(ctx context.Context, orgID *string) ([]domain.Category, error)
    Update(ctx context.Context, category domain.Category) error
    Delete(ctx context.Context, id string) error
}

type categoryRepository struct {
    pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
    return &categoryRepository{pool: pool}
}

func (r *categoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
    query := `
        INSERT INTO categories (name, is_system, org_id, display_order)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, is_system, org_id, display_order, created_at, updated_at
    `

    var result domain.Category
    err := r.pool.QueryRow(ctx, query,
        category.Name,
        category.IsSystem,
        category.OrgID,
        category.DisplayOrder,
    ).Scan(
        &result.ID,
        &result.Name,
        &result.IsSystem,
        &result.OrgID,
        &result.DisplayOrder,
        &result.CreatedAt,
        &result.UpdatedAt,
    )
    if err != nil {
        return domain.Category{}, err
    }

    return result, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (domain.Category, error) {
    query := `
        SELECT id, name, is_system, org_id, display_order, created_at, updated_at
        FROM categories
        WHERE id = $1
    `

    var result domain.Category
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &result.ID,
        &result.Name,
        &result.IsSystem,
        &result.OrgID,
        &result.DisplayOrder,
        &result.CreatedAt,
        &result.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return domain.Category{}, ErrCategoryNotFound
        }
        return domain.Category{}, err
    }

    return result, nil
}

func (r *categoryRepository) List(ctx context.Context, orgID *string) ([]domain.Category, error) {
    query := `
        SELECT id, name, is_system, org_id, display_order, created_at, updated_at
        FROM categories
        WHERE is_system = TRUE OR org_id = $1
        ORDER BY display_order, name
    `

    rows, err := r.pool.Query(ctx, query, orgID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []domain.Category
    for rows.Next() {
        var c domain.Category
        if err := rows.Scan(
            &c.ID,
            &c.Name,
            &c.IsSystem,
            &c.OrgID,
            &c.DisplayOrder,
            &c.CreatedAt,
            &c.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        categories = append(categories, c)
    }

    return categories, rows.Err()
}

func (r *categoryRepository) Update(ctx context.Context, category domain.Category) error {
    query := `
        UPDATE categories
        SET name = $1, display_order = $2, updated_at = NOW()
        WHERE id = $3 AND is_system = FALSE
    `

    result, err := r.pool.Exec(ctx, query, category.Name, category.DisplayOrder, category.ID)
    if err != nil {
        return err
    }

    if result.RowsAffected() == 0 {
        return ErrCategoryNotFound
    }

    return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM categories WHERE id = $1 AND is_system = FALSE`

    result, err := r.pool.Exec(ctx, query, id)
    if err != nil {
        return err
    }

    if result.RowsAffected() == 0 {
        return ErrCategoryNotFound
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./adapters/postgres -run TestCategoryRepository -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/adapters/postgres/category_db.go server/adapters/postgres/category_db_test.go
git commit -m "feat(db): add CategoryRepository"
```

---

### Task 7: Update Rule Repository for New Fields

**Files:**
- Modify: `server/adapters/postgres/rule_db.go`

**Step 1: Update CreateRule method**

In `server/adapters/postgres/rule_db.go`, update the `CreateRule` function:

```go
func (r *ruleRepository) CreateRule(ctx context.Context, rule domain.Rule) (domain.Rule, error) {
    triggersJSON, err := json.Marshal(rule.Triggers)
    if err != nil {
        return domain.Rule{}, err
    }

    query := `
        INSERT INTO rules (
            id, name, content, description, target_layer, category_id,
            priority_weight, overridable, effective_start, effective_end,
            target_teams, target_users, tags, triggers, team_id,
            status, enforcement_mode, temporary_timeout_hours, created_by
        )
        VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10,
            $11, $12, $13, $14, $15,
            $16, $17, $18, $19
        )
        RETURNING id, name, content, description, target_layer, category_id,
            priority_weight, overridable, effective_start, effective_end,
            target_teams, target_users, tags, triggers, team_id, status,
            enforcement_mode, temporary_timeout_hours, created_by,
            submitted_at, approved_at, created_at, updated_at
    `

    var result domain.Rule
    var triggersData []byte

    err = r.pool.QueryRow(ctx, query,
        rule.ID,
        rule.Name,
        rule.Content,
        rule.Description,
        rule.TargetLayer,
        rule.CategoryID,
        rule.PriorityWeight,
        rule.Overridable,
        rule.EffectiveStart,
        rule.EffectiveEnd,
        rule.TargetTeams,
        rule.TargetUsers,
        rule.Tags,
        triggersJSON,
        rule.TeamID,
        rule.Status,
        rule.EnforcementMode,
        rule.TemporaryTimeoutHours,
        rule.CreatedBy,
    ).Scan(
        &result.ID,
        &result.Name,
        &result.Content,
        &result.Description,
        &result.TargetLayer,
        &result.CategoryID,
        &result.PriorityWeight,
        &result.Overridable,
        &result.EffectiveStart,
        &result.EffectiveEnd,
        &result.TargetTeams,
        &result.TargetUsers,
        &result.Tags,
        &triggersData,
        &result.TeamID,
        &result.Status,
        &result.EnforcementMode,
        &result.TemporaryTimeoutHours,
        &result.CreatedBy,
        &result.SubmittedAt,
        &result.ApprovedAt,
        &result.CreatedAt,
        &result.UpdatedAt,
    )
    if err != nil {
        return domain.Rule{}, err
    }

    if err := json.Unmarshal(triggersData, &result.Triggers); err != nil {
        return domain.Rule{}, err
    }

    return result, nil
}
```

**Step 2: Add GetRulesForMerge method for fetching rules by level**

```go
// GetRulesForMerge returns all approved rules for a given target layer, filtered by targeting
func (r *ruleRepository) GetRulesForMerge(ctx context.Context, targetLayer domain.TargetLayer, userID string, teamIDs []string) ([]domain.Rule, error) {
    query := `
        SELECT id, name, content, description, target_layer, category_id,
            priority_weight, overridable, effective_start, effective_end,
            target_teams, target_users, tags, triggers, team_id, status,
            enforcement_mode, temporary_timeout_hours, created_by,
            submitted_at, approved_at, created_at, updated_at
        FROM rules
        WHERE target_layer = $1
          AND status = 'approved'
          AND (
              target_layer = 'enterprise'
              OR (target_teams = '{}' AND target_users = '{}')
              OR target_users @> ARRAY[$2]::uuid[]
              OR target_teams && $3::uuid[]
          )
        ORDER BY priority_weight DESC, name
    `

    rows, err := r.pool.Query(ctx, query, targetLayer, userID, teamIDs)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var rules []domain.Rule
    for rows.Next() {
        var rule domain.Rule
        var triggersData []byte

        if err := rows.Scan(
            &rule.ID,
            &rule.Name,
            &rule.Content,
            &rule.Description,
            &rule.TargetLayer,
            &rule.CategoryID,
            &rule.PriorityWeight,
            &rule.Overridable,
            &rule.EffectiveStart,
            &rule.EffectiveEnd,
            &rule.TargetTeams,
            &rule.TargetUsers,
            &rule.Tags,
            &triggersData,
            &rule.TeamID,
            &rule.Status,
            &rule.EnforcementMode,
            &rule.TemporaryTimeoutHours,
            &rule.CreatedBy,
            &rule.SubmittedAt,
            &rule.ApprovedAt,
            &rule.CreatedAt,
            &rule.UpdatedAt,
        ); err != nil {
            return nil, err
        }

        if err := json.Unmarshal(triggersData, &rule.Triggers); err != nil {
            return nil, err
        }

        rules = append(rules, rule)
    }

    return rules, rows.Err()
}
```

**Step 3: Update scan helper to include new fields**

Update any shared scan functions to handle new fields.

**Step 4: Run existing tests to verify nothing broke**

Run: `cd server && go test ./adapters/postgres/... -v`
Expected: PASS (or identify any tests that need updating)

**Step 5: Commit**

```bash
git add server/adapters/postgres/rule_db.go
git commit -m "feat(db): update RuleRepository for three-file model"
```

---

## Phase 4: Rule Merge Service

### Task 8: Create Merge Service

**Files:**
- Create: `server/services/merge/service.go`
- Create: `server/services/merge/service_test.go`

**Step 1: Write the failing test**

```go
// server/services/merge/service_test.go
package merge

import (
    "testing"

    "claudeception/server/domain"
)

func TestMergeService_RenderManagedSection(t *testing.T) {
    categories := []domain.Category{
        {ID: "cat-1", Name: "Security", DisplayOrder: 1},
        {ID: "cat-2", Name: "Testing", DisplayOrder: 2},
    }

    rules := []domain.Rule{
        {
            Name:        "No Secrets",
            Content:     "Never commit API keys",
            CategoryID:  strPtr("cat-1"),
            TargetLayer: domain.TargetLayerEnterprise,
            Overridable: false,
        },
        {
            Name:        "Min Coverage",
            Content:     "Maintain 80% coverage",
            CategoryID:  strPtr("cat-2"),
            TargetLayer: domain.TargetLayerEnterprise,
            Overridable: true,
        },
    }

    svc := NewService()
    result := svc.RenderManagedSection(rules, categories)

    expected := `<!-- MANAGED BY CLAUDECEPTION - DO NOT EDIT -->

## Security

[Enterprise] **No Secrets**
Never commit API keys

## Testing

[Enterprise] **Min Coverage** (overridable)
Maintain 80% coverage

<!-- END CLAUDECEPTION -->`

    if result != expected {
        t.Errorf("RenderManagedSection() mismatch\ngot:\n%s\n\nwant:\n%s", result, expected)
    }
}

func strPtr(s string) *string {
    return &s
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./services/merge -run TestMergeService -v`
Expected: FAIL - package/types not defined

**Step 3: Write the implementation**

```go
// server/services/merge/service.go
package merge

import (
    "fmt"
    "sort"
    "strings"

    "claudeception/server/domain"
)

const (
    ManagedSectionStart = "<!-- MANAGED BY CLAUDECEPTION - DO NOT EDIT -->"
    ManagedSectionEnd   = "<!-- END CLAUDECEPTION -->"
)

type Service struct{}

func NewService() *Service {
    return &Service{}
}

// RenderManagedSection generates the managed CLAUDE.md section from rules
func (s *Service) RenderManagedSection(rules []domain.Rule, categories []domain.Category) string {
    if len(rules) == 0 {
        return ""
    }

    // Build category lookup
    categoryMap := make(map[string]domain.Category)
    for _, c := range categories {
        categoryMap[c.ID] = c
    }

    // Group rules by category
    rulesByCategory := make(map[string][]domain.Rule)
    for _, r := range rules {
        catID := ""
        if r.CategoryID != nil {
            catID = *r.CategoryID
        }
        rulesByCategory[catID] = append(rulesByCategory[catID], r)
    }

    // Sort categories by display order
    var sortedCatIDs []string
    for catID := range rulesByCategory {
        sortedCatIDs = append(sortedCatIDs, catID)
    }
    sort.Slice(sortedCatIDs, func(i, j int) bool {
        catI := categoryMap[sortedCatIDs[i]]
        catJ := categoryMap[sortedCatIDs[j]]
        if catI.DisplayOrder != catJ.DisplayOrder {
            return catI.DisplayOrder < catJ.DisplayOrder
        }
        return catI.Name < catJ.Name
    })

    var sections []string
    sections = append(sections, ManagedSectionStart)

    for _, catID := range sortedCatIDs {
        rules := rulesByCategory[catID]
        cat := categoryMap[catID]

        // Sort rules by priority within category
        sort.Slice(rules, func(i, j int) bool {
            return rules[i].PriorityWeight > rules[j].PriorityWeight
        })

        sections = append(sections, fmt.Sprintf("\n## %s\n", cat.Name))

        for _, r := range rules {
            levelTag := fmt.Sprintf("[%s]", strings.Title(string(r.TargetLayer)))
            overridableTag := ""
            if r.Overridable {
                overridableTag = " (overridable)"
            }

            sections = append(sections, fmt.Sprintf("%s **%s**%s\n%s", levelTag, r.Name, overridableTag, r.Content))
        }
    }

    sections = append(sections, "\n"+ManagedSectionEnd)

    return strings.Join(sections, "\n")
}

// MergeWithExisting combines managed section with existing file content
func (s *Service) MergeWithExisting(existingContent, managedSection string) string {
    startIdx := strings.Index(existingContent, ManagedSectionStart)
    endIdx := strings.Index(existingContent, ManagedSectionEnd)

    if startIdx == -1 {
        // No existing managed section - append at end
        if existingContent != "" && !strings.HasSuffix(existingContent, "\n\n") {
            existingContent = strings.TrimRight(existingContent, "\n") + "\n\n"
        }
        return existingContent + managedSection
    }

    // Replace existing managed section
    before := existingContent[:startIdx]
    after := ""
    if endIdx != -1 {
        after = existingContent[endIdx+len(ManagedSectionEnd):]
    }

    return before + managedSection + after
}

// ExtractManualContent returns content outside the managed section
func (s *Service) ExtractManualContent(content string) (before, after string) {
    startIdx := strings.Index(content, ManagedSectionStart)
    endIdx := strings.Index(content, ManagedSectionEnd)

    if startIdx == -1 {
        return content, ""
    }

    before = content[:startIdx]
    if endIdx != -1 {
        after = content[endIdx+len(ManagedSectionEnd):]
    }

    return before, after
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./services/merge -run TestMergeService -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/services/merge/service.go server/services/merge/service_test.go
git commit -m "feat(services): add merge service for rendering managed sections"
```

---

## Phase 5: Category API Handlers

### Task 9: Create Category Handler

**Files:**
- Create: `server/entrypoints/api/handlers/categories.go`
- Create: `server/entrypoints/api/handlers/categories_test.go`

**Step 1: Write the failing test**

```go
// server/entrypoints/api/handlers/categories_test.go
package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "claudeception/server/domain"

    "github.com/go-chi/chi/v5"
)

type mockCategoryService struct {
    categories map[string]domain.Category
}

func (m *mockCategoryService) Create(ctx context.Context, req CreateCategoryRequest) (domain.Category, error) {
    cat := domain.Category{
        ID:           "cat-new",
        Name:         req.Name,
        IsSystem:     false,
        OrgID:        req.OrgID,
        DisplayOrder: req.DisplayOrder,
    }
    m.categories[cat.ID] = cat
    return cat, nil
}

func (m *mockCategoryService) List(ctx context.Context, orgID *string) ([]domain.Category, error) {
    var result []domain.Category
    for _, c := range m.categories {
        result = append(result, c)
    }
    return result, nil
}

func (m *mockCategoryService) Delete(ctx context.Context, id string) error {
    delete(m.categories, id)
    return nil
}

func TestCategoriesHandler_List(t *testing.T) {
    mock := &mockCategoryService{
        categories: map[string]domain.Category{
            "cat-1": {ID: "cat-1", Name: "Security", IsSystem: true},
        },
    }
    h := NewCategoriesHandler(mock)

    req := httptest.NewRequest(http.MethodGet, "/categories", nil)
    rec := httptest.NewRecorder()

    h.List(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", rec.Code)
    }
}

func TestCategoriesHandler_Create(t *testing.T) {
    mock := &mockCategoryService{categories: make(map[string]domain.Category)}
    h := NewCategoriesHandler(mock)

    body := `{"name": "Custom Category", "display_order": 5}`
    req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    h.Create(rec, req)

    if rec.Code != http.StatusCreated {
        t.Errorf("expected status 201, got %d", rec.Code)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./entrypoints/api/handlers -run TestCategoriesHandler -v`
Expected: FAIL - types not defined

**Step 3: Write the implementation**

```go
// server/entrypoints/api/handlers/categories.go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"

    "claudeception/server/domain"

    "github.com/go-chi/chi/v5"
)

type CategoryService interface {
    Create(ctx context.Context, req CreateCategoryRequest) (domain.Category, error)
    List(ctx context.Context, orgID *string) ([]domain.Category, error)
    Delete(ctx context.Context, id string) error
}

type CategoriesHandler struct {
    service CategoryService
}

func NewCategoriesHandler(service CategoryService) *CategoriesHandler {
    return &CategoriesHandler{service: service}
}

func (h *CategoriesHandler) RegisterRoutes(r chi.Router) {
    r.Get("/", h.List)
    r.Post("/", h.Create)
    r.Delete("/{id}", h.Delete)
}

type CreateCategoryRequest struct {
    Name         string  `json:"name"`
    OrgID        *string `json:"org_id,omitempty"`
    DisplayOrder int     `json:"display_order"`
}

func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
    orgID := r.URL.Query().Get("org_id")
    var orgIDPtr *string
    if orgID != "" {
        orgIDPtr = &orgID
    }

    categories, err := h.service.List(r.Context(), orgIDPtr)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(categories)
}

func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateCategoryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    category, err := h.service.Create(r.Context(), req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(category)
}

func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    if err := h.service.Delete(r.Context(), id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./entrypoints/api/handlers -run TestCategoriesHandler -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/entrypoints/api/handlers/categories.go server/entrypoints/api/handlers/categories_test.go
git commit -m "feat(api): add categories handler"
```

---

### Task 10: Register Category Routes

**Files:**
- Modify: `server/entrypoints/api/router.go`

**Step 1: Add categories route**

Add to `router.go` after rules route:

```go
r.Route("/categories", func(r chi.Router) {
    h := handlers.NewCategoriesHandler(cfg.CategoryService)
    h.RegisterRoutes(r)
})
```

**Step 2: Update config type to include CategoryService**

Modify the router config struct to include `CategoryService`.

**Step 3: Run build to verify compilation**

Run: `cd server && go build ./...`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add server/entrypoints/api/router.go
git commit -m "feat(api): register categories routes"
```

---

## Phase 6: Update Rules Handler

### Task 11: Add Merged Content Endpoint

**Files:**
- Modify: `server/entrypoints/api/handlers/rules.go`

**Step 1: Add GetMerged handler**

Add to `rules.go`:

```go
func (h *RulesHandler) GetMerged(w http.ResponseWriter, r *http.Request) {
    level := r.URL.Query().Get("level")
    if level == "" {
        http.Error(w, "level query parameter required", http.StatusBadRequest)
        return
    }

    userID := r.Context().Value("user_id").(string)
    teamIDs := r.Context().Value("team_ids").([]string)

    content, err := h.service.GetMergedContent(r.Context(), domain.TargetLayer(level), userID, teamIDs)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/markdown")
    w.Write([]byte(content))
}
```

**Step 2: Register the route**

Update `RegisterRoutes` in `rules.go`:

```go
r.Get("/merged", h.GetMerged)
```

**Step 3: Run build to verify**

Run: `cd server && go build ./...`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add server/entrypoints/api/handlers/rules.go
git commit -m "feat(api): add merged content endpoint"
```

---

## Phase 7: Agent Changes

### Task 12: Update Agent Storage Schema

**Files:**
- Modify: `agent/storage/migrations.go`
- Modify: `agent/storage/rules.go`

**Step 1: Update migration to add new columns**

Add new migration in `migrations.go`:

```go
const migrationV2 = `
ALTER TABLE cached_rules ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
ALTER TABLE cached_rules ADD COLUMN IF NOT EXISTS category_id TEXT;
ALTER TABLE cached_rules ADD COLUMN IF NOT EXISTS category_name TEXT;
ALTER TABLE cached_rules ADD COLUMN IF NOT EXISTS overridable INTEGER DEFAULT 1;
ALTER TABLE cached_rules ADD COLUMN IF NOT EXISTS effective_start INTEGER;
ALTER TABLE cached_rules ADD COLUMN IF NOT EXISTS effective_end INTEGER;
ALTER TABLE cached_rules ADD COLUMN IF NOT EXISTS tags TEXT DEFAULT '[]';
`
```

**Step 2: Update CachedRule struct**

```go
type CachedRule struct {
    ID                    string          `json:"id"`
    Name                  string          `json:"name"`
    Content               string          `json:"content"`
    Description           string          `json:"description"`
    TargetLayer           string          `json:"target_layer"`
    CategoryID            string          `json:"category_id"`
    CategoryName          string          `json:"category_name"`
    Overridable           bool            `json:"overridable"`
    EffectiveStart        *int64          `json:"effective_start"`
    EffectiveEnd          *int64          `json:"effective_end"`
    Tags                  json.RawMessage `json:"tags"`
    Triggers              json.RawMessage `json:"triggers"`
    EnforcementMode       string          `json:"enforcement_mode"`
    TemporaryTimeoutHours int             `json:"temporary_timeout_hours"`
    Version               int             `json:"version"`
    CachedAt              time.Time       `json:"cached_at"`
}
```

**Step 3: Update save/get methods for new fields**

**Step 4: Run agent tests**

Run: `cd agent && go test ./storage/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/storage/migrations.go agent/storage/rules.go
git commit -m "feat(agent): update storage schema for three-file model"
```

---

### Task 13: Create Merge Renderer in Agent

**Files:**
- Create: `agent/renderer/renderer.go`
- Create: `agent/renderer/renderer_test.go`

**Step 1: Write the failing test**

```go
// agent/renderer/renderer_test.go
package renderer

import (
    "testing"

    "claudeception/agent/storage"
)

func TestRenderer_RenderManagedSection(t *testing.T) {
    rules := []storage.CachedRule{
        {
            Name:         "No Secrets",
            Content:      "Never commit API keys",
            CategoryName: "Security",
            TargetLayer:  "enterprise",
            Overridable:  false,
        },
    }

    r := New()
    result := r.RenderManagedSection(rules)

    if result == "" {
        t.Error("expected non-empty result")
    }

    if !strings.Contains(result, "<!-- MANAGED BY CLAUDECEPTION") {
        t.Error("expected managed section markers")
    }

    if !strings.Contains(result, "No Secrets") {
        t.Error("expected rule name in output")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./renderer -v`
Expected: FAIL - package not defined

**Step 3: Write the implementation**

```go
// agent/renderer/renderer.go
package renderer

import (
    "fmt"
    "sort"
    "strings"
    "time"

    "claudeception/agent/storage"
)

const (
    ManagedSectionStart = "<!-- MANAGED BY CLAUDECEPTION - DO NOT EDIT -->"
    ManagedSectionEnd   = "<!-- END CLAUDECEPTION -->"
)

type Renderer struct{}

func New() *Renderer {
    return &Renderer{}
}

// RenderManagedSection generates the managed content from cached rules
func (r *Renderer) RenderManagedSection(rules []storage.CachedRule) string {
    // Filter by effective dates
    now := time.Now().Unix()
    var activeRules []storage.CachedRule
    for _, rule := range rules {
        if rule.EffectiveStart != nil && now < *rule.EffectiveStart {
            continue
        }
        if rule.EffectiveEnd != nil && now > *rule.EffectiveEnd {
            continue
        }
        activeRules = append(activeRules, rule)
    }

    if len(activeRules) == 0 {
        return ""
    }

    // Group by category
    byCategory := make(map[string][]storage.CachedRule)
    for _, rule := range activeRules {
        cat := rule.CategoryName
        if cat == "" {
            cat = "Uncategorized"
        }
        byCategory[cat] = append(byCategory[cat], rule)
    }

    // Sort categories
    var categories []string
    for cat := range byCategory {
        categories = append(categories, cat)
    }
    sort.Strings(categories)

    var sections []string
    sections = append(sections, ManagedSectionStart)

    for _, cat := range categories {
        catRules := byCategory[cat]
        sections = append(sections, fmt.Sprintf("\n## %s\n", cat))

        for _, rule := range catRules {
            levelTag := fmt.Sprintf("[%s]", strings.Title(rule.TargetLayer))
            overridableTag := ""
            if rule.Overridable {
                overridableTag = " (overridable)"
            }
            sections = append(sections, fmt.Sprintf("%s **%s**%s\n%s", levelTag, rule.Name, overridableTag, rule.Content))
        }
    }

    sections = append(sections, "\n"+ManagedSectionEnd)
    return strings.Join(sections, "\n")
}

// MergeWithFile combines managed section with existing file content
func (r *Renderer) MergeWithFile(existing, managed string) string {
    startIdx := strings.Index(existing, ManagedSectionStart)
    endIdx := strings.Index(existing, ManagedSectionEnd)

    if startIdx == -1 {
        if existing != "" && !strings.HasSuffix(existing, "\n\n") {
            existing = strings.TrimRight(existing, "\n") + "\n\n"
        }
        return existing + managed
    }

    before := existing[:startIdx]
    after := ""
    if endIdx != -1 {
        after = existing[endIdx+len(ManagedSectionEnd):]
    }

    return before + managed + after
}

// DetectManagedSectionTampering checks if managed section was modified
func (r *Renderer) DetectManagedSectionTampering(fileContent, expectedManaged string) bool {
    startIdx := strings.Index(fileContent, ManagedSectionStart)
    endIdx := strings.Index(fileContent, ManagedSectionEnd)

    if startIdx == -1 || endIdx == -1 {
        return expectedManaged != ""
    }

    actual := fileContent[startIdx : endIdx+len(ManagedSectionEnd)]
    return actual != expectedManaged
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./renderer -v`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/renderer/renderer.go agent/renderer/renderer_test.go
git commit -m "feat(agent): add merge renderer for CLAUDE.md files"
```

---

### Task 14: Update Daemon for Three Fixed Paths

**Files:**
- Modify: `agent/daemon/daemon.go`

**Step 1: Define fixed paths**

Add constants:

```go
const (
    EnterpriseFilePath = "/etc/claude-code/CLAUDE.md"
    UserFilePath       = "~/.claude/CLAUDE.md"  // Expand at runtime
    ProjectFileName    = "CLAUDE.md"
)

type ManagedFile struct {
    Level string
    Path  string
}
```

**Step 2: Update daemon to track three paths**

Update `Daemon` struct:

```go
type Daemon struct {
    store         *storage.Storage
    wsClient      *ws.Client
    serverURL     string
    listener      net.Listener
    fileWatcher   *watcher.Watcher
    renderer      *renderer.Renderer
    managedFiles  map[string]ManagedFile  // path -> level
    projectDirs   []string                // watched project directories
}
```

**Step 3: Add sync method**

```go
func (d *Daemon) syncFile(level string, path string) error {
    rules, err := d.store.GetRulesByLevel(level)
    if err != nil {
        return err
    }

    managed := d.renderer.RenderManagedSection(rules)

    existing, err := os.ReadFile(path)
    if err != nil && !os.IsNotExist(err) {
        return err
    }

    merged := d.renderer.MergeWithFile(string(existing), managed)

    return os.WriteFile(path, []byte(merged), 0644)
}
```

**Step 4: Run daemon tests**

Run: `cd agent && go test ./daemon/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/daemon/daemon.go
git commit -m "feat(agent): update daemon for three fixed paths"
```

---

## Phase 8: WebUI Changes

### Task 15: Update Rule TypeScript Types

**Files:**
- Modify: `web/src/domain/rule.ts`

**Step 1: Update Rule interface**

```typescript
export type TargetLayer = 'enterprise' | 'user' | 'project';

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
  teamId: string;
  status: RuleStatus;
  enforcementMode: EnforcementMode;
  temporaryTimeoutHours: number;
  createdBy?: string;
  submittedAt?: string;
  approvedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Category {
  id: string;
  name: string;
  isSystem: boolean;
  orgId?: string;
  displayOrder: number;
}
```

**Step 2: Commit**

```bash
git add web/src/domain/rule.ts
git commit -m "feat(web): update Rule types for three-file model"
```

---

### Task 16: Create Category API Client

**Files:**
- Modify: `web/src/lib/api.ts`

**Step 1: Add category API functions**

```typescript
export async function listCategories(orgId?: string): Promise<Category[]> {
  const params = orgId ? `?org_id=${orgId}` : '';
  const response = await fetch(`${API_BASE}/categories${params}`, {
    headers: getAuthHeaders(),
  });
  return response.json();
}

export async function createCategory(data: {
  name: string;
  orgId?: string;
  displayOrder: number;
}): Promise<Category> {
  const response = await fetch(`${API_BASE}/categories`, {
    method: 'POST',
    headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  return response.json();
}

export async function deleteCategory(id: string): Promise<void> {
  await fetch(`${API_BASE}/categories/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
}

export async function getMergedContent(level: TargetLayer): Promise<string> {
  const response = await fetch(`${API_BASE}/rules/merged?level=${level}`, {
    headers: getAuthHeaders(),
  });
  return response.text();
}
```

**Step 2: Commit**

```bash
git add web/src/lib/api.ts
git commit -m "feat(web): add category API client"
```

---

### Task 17: Update RuleEditor Form

**Files:**
- Modify: `web/src/components/RuleEditor.tsx`

**Step 1: Add new form fields**

Update the form to include:
- Description (textarea)
- Category (dropdown from categories API)
- Overridable (checkbox)
- Effective Start/End (date pickers)
- Target Teams (multi-select, hidden for enterprise)
- Target Users (multi-select, hidden for enterprise)
- Tags (tag input)

**Step 2: Add conditional visibility**

```typescript
{formData.targetLayer !== 'enterprise' && (
  <>
    <div>
      <label>Target Teams</label>
      <MultiSelect
        options={teams}
        value={formData.targetTeams}
        onChange={(v) => setFormData({...formData, targetTeams: v})}
      />
    </div>
    <div>
      <label>Target Users</label>
      <MultiSelect
        options={users}
        value={formData.targetUsers}
        onChange={(v) => setFormData({...formData, targetUsers: v})}
      />
    </div>
  </>
)}
```

**Step 3: Test the form renders**

Run: `cd web && npm run dev`
Navigate to rule editor, verify new fields appear

**Step 4: Commit**

```bash
git add web/src/components/RuleEditor.tsx
git commit -m "feat(web): update RuleEditor with three-file model fields"
```

---

### Task 18: Create Category Management Page

**Files:**
- Create: `web/src/app/settings/categories/page.tsx`

**Step 1: Create the page**

```typescript
'use client';

import { useState, useEffect } from 'react';
import { Category, listCategories, createCategory, deleteCategory } from '@/lib/api';

export default function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [newName, setNewName] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadCategories();
  }, []);

  async function loadCategories() {
    const cats = await listCategories();
    setCategories(cats);
    setLoading(false);
  }

  async function handleCreate() {
    if (!newName.trim()) return;
    await createCategory({
      name: newName,
      displayOrder: categories.length + 1,
    });
    setNewName('');
    loadCategories();
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this category?')) return;
    await deleteCategory(id);
    loadCategories();
  }

  if (loading) return <div>Loading...</div>;

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Categories</h1>

      <div className="mb-4 flex gap-2">
        <input
          type="text"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder="New category name"
          className="border p-2 rounded"
        />
        <button onClick={handleCreate} className="bg-blue-500 text-white px-4 py-2 rounded">
          Add
        </button>
      </div>

      <ul className="space-y-2">
        {categories.map((cat) => (
          <li key={cat.id} className="flex justify-between items-center p-2 border rounded">
            <span>
              {cat.name}
              {cat.isSystem && <span className="ml-2 text-gray-500">(System)</span>}
            </span>
            {!cat.isSystem && (
              <button
                onClick={() => handleDelete(cat.id)}
                className="text-red-500"
              >
                Delete
              </button>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

**Step 2: Run dev server and test**

Run: `cd web && npm run dev`
Navigate to /settings/categories, verify page works

**Step 3: Commit**

```bash
git add web/src/app/settings/categories/page.tsx
git commit -m "feat(web): add category management page"
```

---

## Phase 9: Integration Testing

### Task 19: Write Integration Tests for Rule Merging

**Files:**
- Create: `server/integration/merge_test.go`

**Step 1: Write integration test**

```go
//go:build integration

package integration

import (
    "context"
    "testing"

    "claudeception/server/domain"
    "claudeception/server/services/merge"
)

func TestMergeService_Integration(t *testing.T) {
    resetDB(t)

    // Create categories
    cat1 := testFixtures.CreateCategory(t, "Security", true)
    cat2 := testFixtures.CreateCategory(t, "Testing", true)

    // Create rules at different levels
    rule1 := testFixtures.CreateRule(t, domain.Rule{
        Name:        "No Secrets",
        Content:     "Never commit API keys",
        TargetLayer: domain.TargetLayerEnterprise,
        CategoryID:  &cat1.ID,
        Overridable: false,
    })

    rule2 := testFixtures.CreateRule(t, domain.Rule{
        Name:        "Coverage",
        Content:     "80% minimum",
        TargetLayer: domain.TargetLayerEnterprise,
        CategoryID:  &cat2.ID,
        Overridable: true,
    })

    // Get merged content
    svc := merge.NewService()
    categories := []domain.Category{cat1, cat2}
    rules := []domain.Rule{rule1, rule2}

    result := svc.RenderManagedSection(rules, categories)

    if result == "" {
        t.Error("expected non-empty result")
    }

    t.Log(result)
}
```

**Step 2: Run integration test**

Run: `cd server && go test -tags=integration ./integration -run TestMergeService -v`
Expected: PASS

**Step 3: Commit**

```bash
git add server/integration/merge_test.go
git commit -m "test: add merge service integration test"
```

---

### Task 20: Final Build and Verification

**Step 1: Build all components**

Run: `task build` (or equivalent)
Expected: All components build successfully

**Step 2: Run all tests**

Run: `task test` (or equivalent)
Expected: All tests pass

**Step 3: Create final commit**

```bash
git add -A
git commit -m "feat: complete three-file model implementation"
```

---

## Summary

This implementation transforms Claudeception to manage exactly three CLAUDE.md files with the following key changes:

1. **Database**: New categories table, extended rules table with new fields
2. **Domain**: Updated Rule model with override validation, effective date checking
3. **Services**: New merge service for rendering managed sections
4. **API**: Category CRUD endpoints, merged content endpoint
5. **Agent**: New renderer, updated daemon for fixed paths, managed section detection
6. **WebUI**: Updated rule form, category management page

Total: 20 tasks across 9 phases
