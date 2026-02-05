# Global Rules Design

## Overview

This design adds organization-wide global rules that exist outside team ownership, with selective inheritance and forced enforcement capabilities.

## Features

- **Global rules** - Rules with no team ownership (`team_id = NULL`), created by any admin
- **Selective inheritance** - Each global rule has a `force` flag
- **Team settings** - Teams have `inherit_global_rules` setting (default: true)
- **Force overrides opt-out** - Global rules with `force: true` apply to ALL teams regardless of their inheritance setting
- **Visible enforcement** - Forced rules are clearly marked in the UI

## Data Model Changes

### Rule Model

Add to `domain.Rule`:

```go
Force bool `json:"force"` // Only valid for global rules (team_id = NULL)
```

Schema change: `team_id` becomes **nullable**. When `team_id IS NULL`, the rule is global.

Constraints:
- `force` can only be `true` when `team_id IS NULL`
- Global rules must have `target_layer = 'enterprise'`
- Global rules ignore `target_teams` and `target_users` (they apply org-wide)

### Team Model

Add to `domain.TeamSettings`:

```go
type TeamSettings struct {
    DriftThresholdMinutes int  `json:"drift_threshold_minutes"`
    InheritGlobalRules    bool `json:"inherit_global_rules"` // default: true
}
```

## Query Logic for Rule Application

### Which global rules apply to a team?

```
IF global_rule.force = true:
    → Always applies (ignores team settings)
ELSE IF team.settings.inherit_global_rules = true:
    → Applies
ELSE:
    → Does not apply
```

### Updated GetRulesForMerge Query Logic

```sql
-- Global rules (team_id IS NULL)
WHERE team_id IS NULL
  AND status = 'approved'
  AND (
      force = true                                    -- Forced rules always apply
      OR (SELECT inherit_global_rules FROM teams WHERE id = $team_id)  -- Team opted in
  )

UNION ALL

-- Team-owned rules (existing logic)
WHERE team_id IS NOT NULL
  AND status = 'approved'
  AND (
      team_id = $team_id                              -- Owned by this team
      OR $team_id = ANY(target_teams)                 -- Cross-team targeting
      OR $user_id = ANY(target_users)                 -- User targeting
  )
```

### Priority Ordering

Global rules appear first (highest priority), then team rules by their `target_layer` and `priority_weight`.

## API & Permission Changes

### Rule Creation

**POST `/api/rules`** - Updated validation:

| Condition | Behavior |
|-----------|----------|
| `team_id` omitted/null | Creates global rule (admin required) |
| `team_id` provided | Creates team-scoped rule (existing behavior) |
| `force: true` + `team_id` provided | Rejected with error |
| `force: true` + non-admin | Rejected (admins only) |

**Request body** (updated):

```json
{
  "name": "No Hardcoded Secrets",
  "content": "Never commit API keys...",
  "target_layer": "enterprise",
  "team_id": null,
  "force": true
}
```

### Team Settings

**PATCH `/api/teams/:id/settings`** - Update inheritance:

```json
{
  "inherit_global_rules": false
}
```

Only team admins or super admins can change this setting.

### New Endpoint

**GET `/api/rules/global`** - List all global rules (admins only)

Returns rules where `team_id IS NULL`, with `force` status visible.

## UI Changes

### Rule List View

**New "Global" tab** alongside existing team-scoped rules:
- Shows all rules where `team_id IS NULL`
- Column: "Enforcement" showing "Forced" or "Inheritable"
- Only visible to admins

### Rule Creation Form

**When admin selects "Global" scope:**
- `team_id` field hidden/disabled
- New checkbox: "Force on all teams" (default: unchecked)
- Tooltip: "Forced rules apply even to teams that opted out of global rule inheritance"

### Team Settings Page

**New toggle in team settings:**

```
Inherit Global Rules: [ON/OFF]
"When disabled, this team will only receive forced global rules"
```

Show count of forced rules: "3 forced rules will always apply"

### Rule Display (in merged view)

| Type | Display |
|------|---------|
| Forced global rule | `[Forced Global] **Rule Name**` |
| Inherited global rule | `[Global] **Rule Name**` |
| Team rule | `[Team] **Rule Name**` |

## Database Migration

```sql
-- 1. Make team_id nullable
ALTER TABLE rules ALTER COLUMN team_id DROP NOT NULL;

-- 2. Add force column
ALTER TABLE rules ADD COLUMN force BOOLEAN NOT NULL DEFAULT false;

-- 3. Add constraint: force only valid for global rules
ALTER TABLE rules ADD CONSTRAINT rules_force_global_only
    CHECK (force = false OR team_id IS NULL);

-- 4. Add constraint: global rules must be enterprise layer
ALTER TABLE rules ADD CONSTRAINT rules_global_enterprise_only
    CHECK (team_id IS NOT NULL OR target_layer = 'enterprise');

-- 5. Index for global rules
CREATE INDEX idx_rules_global ON rules(force) WHERE team_id IS NULL;
```

## Backwards Compatibility

- **Existing rules unchanged** - All have `team_id`, `force = false`
- **Existing queries work** - `team_id IS NOT NULL` filters still function
- **API accepts both** - `team_id` can be provided (team rule) or omitted (global rule)
- **Agent sync** - Updated query returns global + team rules; agent just renders them

## Implementation Scope

### Server (Domain & DB)

- `server/domain/rule.go` - Add `Force` field, validation methods
- `server/domain/team.go` - Add `InheritGlobalRules` to TeamSettings
- `server/migrations/000023_add_global_rules.up.sql` - Schema changes
- `server/adapters/postgres/rule_db.go` - Update queries for global rules
- `server/adapters/postgres/team_db.go` - Handle new team setting

### Server (API)

- `server/entrypoints/api/handlers/rules.go` - Validate global rule creation, add `/rules/global` endpoint
- `server/entrypoints/api/handlers/teams.go` - Handle settings update
- `server/services/rules/repository.go` - Update interface
- `server/services/merge/service.go` - Include global rules in merge

### Web

- `web/src/domain/rule.ts` - Add `force` field
- `web/src/components/RuleEditor.tsx` - Global scope option, force checkbox
- `web/src/components/RuleList.tsx` - Global tab, enforcement badges
- `web/src/app/admin/teams/[id]/settings/page.tsx` - Inheritance toggle

### Agent

- No changes needed - agent receives merged rules via existing sync
