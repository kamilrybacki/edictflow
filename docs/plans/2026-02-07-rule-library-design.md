# Rule Library Design

## Overview

This document describes a redesign of how rules are organized in Edictflow. Instead of rules being owned by teams, all rules will live in a central **Rule Library**. Administrators can then **attach** library rules to teams with per-attachment enforcement settings.

## Goals

- Centralize rule management in a single library
- Enable rule reuse across teams without duplication
- Provide granular control over how rules are enforced per team
- Maintain governance through dual approval (library entry + attachment)

## Core Concepts

### Rule Library

The Rule Library is the single source of truth for all rules in Edictflow. Rules exist independently of teams and are attached to teams as needed.

**Library rule lifecycle:**

1. Team admin or org admin creates a rule → status: `draft`
2. Submits for approval → status: `pending`
3. Org admin approves → status: `approved` (now attachable)
4. Org admin rejects → status: `rejected` (back to author for revision)

**Key properties of a library rule:**

- Content, name, description, category, triggers, priority, tags
- `target_layer`: Enterprise, Team, or Project
- `created_by`: The user who authored it
- `approved_by`: The org admin who approved it
- No `team_id` - rules are team-agnostic in the library

**Enterprise rules** are special: once approved, they are automatically attached to ALL teams with `approved` status. Other rules require explicit attachment.

### Rule Attachments

When an admin attaches a library rule to a team, a `RuleAttachment` record is created.

**Attachment properties:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique identifier |
| `rule_id` | UUID | Reference to the library rule |
| `team_id` | UUID | Team this attachment applies to |
| `enforcement_mode` | Enum | block, warning, or temporary |
| `temporary_timeout_hours` | Integer | Timeout when mode is temporary |
| `status` | Enum | pending, approved, rejected |
| `requested_by` | UUID | Admin who requested the attachment |
| `approved_by` | UUID | Org admin who approved (nullable) |
| `created_at` | Timestamp | When attachment was requested |
| `approved_at` | Timestamp | When attachment was approved |

**Attachment workflow:**

1. Team admin or org admin requests attachment → status: `pending`
2. Org admin approves → status: `approved` (rule now active for that team)
3. Org admin rejects → status: `rejected`

**Constraints:**

- One attachment per rule-team pair
- To change enforcement mode, update the existing attachment

### Behavior by Rule Type

| Rule Type | Auto-attached? | Approval needed to attach? | Enforcement configurable per team? |
|-----------|----------------|---------------------------|-----------------------------------|
| Enterprise | Yes, to all teams | No (auto on rule approval) | Yes |
| Team | No | Yes | Yes |
| Project | No | Yes | Yes |

## Data Model Changes

### Modified: Rule Entity

Remove team ownership, add library metadata:

```go
type Rule struct {
    ID                    string
    Name                  string
    Content               string
    Description           *string
    TargetLayer           TargetLayer
    CategoryID            *string
    PriorityWeight        int
    Overridable           bool
    EffectiveStart        *time.Time
    EffectiveEnd          *time.Time
    Tags                  []string
    Triggers              []Trigger
    Status                RuleStatus
    CreatedBy             *string
    SubmittedAt           *time.Time
    ApprovedAt            *time.Time
    ApprovedBy            *string         // NEW: org admin who approved
    CreatedAt             time.Time
    UpdatedAt             time.Time
}
```

**Removed fields:**

- `TeamID` - rules are library-wide
- `Force` - replaced by Enterprise layer auto-attachment
- `EnforcementMode` - moved to attachment
- `TemporaryTimeoutHours` - moved to attachment
- `TargetTeams`, `TargetUsers` - replaced by attachments

### New: RuleAttachment Entity

```go
type RuleAttachment struct {
    ID                    string
    RuleID                string
    TeamID                string
    EnforcementMode       EnforcementMode
    TemporaryTimeoutHours int
    Status                AttachmentStatus // pending, approved, rejected
    RequestedBy           string
    ApprovedBy            *string
    CreatedAt             time.Time
    ApprovedAt            *time.Time
}
```

## API Changes

### Library Endpoints (New)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/library/rules` | List all library rules (filterable by status, category, layer) |
| `POST` | `/api/v1/library/rules` | Create a new library rule (draft) |
| `GET` | `/api/v1/library/rules/:id` | Get a single library rule |
| `PUT` | `/api/v1/library/rules/:id` | Update a library rule (draft/rejected only) |
| `DELETE` | `/api/v1/library/rules/:id` | Delete a library rule (draft only, or org admin) |
| `POST` | `/api/v1/library/rules/:id/submit` | Submit for approval |
| `POST` | `/api/v1/library/rules/:id/approve` | Org admin approves (auto-attaches if Enterprise) |
| `POST` | `/api/v1/library/rules/:id/reject` | Org admin rejects |

### Attachment Endpoints (New)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/teams/:teamId/attachments` | List attachments for a team |
| `POST` | `/api/v1/teams/:teamId/attachments` | Request attachment of a rule |
| `GET` | `/api/v1/attachments/:id` | Get attachment details |
| `PUT` | `/api/v1/attachments/:id` | Update enforcement mode |
| `DELETE` | `/api/v1/attachments/:id` | Remove attachment |
| `POST` | `/api/v1/attachments/:id/approve` | Org admin approves attachment |
| `POST` | `/api/v1/attachments/:id/reject` | Org admin rejects attachment |

### Deprecated Endpoints

| Endpoint | Replacement |
|----------|-------------|
| `POST /api/v1/rules` | `POST /api/v1/library/rules` |
| `GET /api/v1/teams/:teamId/rules` | `GET /api/v1/teams/:teamId/attachments` (join with rules) |
| `POST /api/v1/global-rules` | `POST /api/v1/library/rules` with `target_layer: enterprise` |

## UI Changes

### New Views

1. **Rule Library** (`/library`)
   - Browsable list of all library rules
   - Filter by: status (draft/pending/approved), category, target layer, tags
   - Search by name/content
   - Shows attachment count per rule
   - Actions: View, Edit (if draft), Submit, Approve/Reject (org admin)

2. **Library Rule Detail** (`/library/:ruleId`)
   - Full rule content with markdown preview
   - List of teams this rule is attached to
   - Attachment status per team
   - "Attach to Team" button

3. **Team Attachments** (`/teams/:teamId/attachments`)
   - List of rules attached to this team
   - Shows enforcement mode, status per attachment
   - Actions: Change enforcement, Detach, Approve/Reject pending
   - "Browse Library" button to attach new rules

### Modified Views

1. **Rule Editor** - Simplified
   - Removes team selector, enforcement mode, force flag
   - Focuses on rule content: name, description, content, category, triggers, tags, layer
   - Used for both creating and editing library rules

2. **Attachment Modal** - New
   - Select enforcement mode (block/warning/temporary)
   - If temporary, set timeout hours
   - Submit for approval

### Removed

- "Create Rule" from team dashboard (replaced by "Browse Library" → "Attach")
- Global rule toggle in editor

## Migration Strategy

### Database Migration

1. **Create `rule_attachments` table**
   - Schema as defined above

2. **Migrate existing rules:**
   - For each existing rule with `team_id`:
     - Set `team_id = NULL` on the rule (move to library)
     - Create a `RuleAttachment` linking rule → original team
     - Copy `enforcement_mode` and `temporary_timeout_hours` to attachment
     - Set attachment status = `approved` (preserve existing behavior)

3. **Migrate global/forced rules:**
   - Rules with `force = true` or `target_layer = enterprise`:
     - Create `approved` attachments for ALL existing teams
     - Copy enforcement mode to each attachment

4. **Clean up Rule table:**
   - Remove columns: `team_id`, `force`, `enforcement_mode`, `temporary_timeout_hours`, `target_teams`, `target_users`
   - Add column: `approved_by`

### Rollback Plan

- Keep old columns nullable during transition
- Migration script stores original `team_id` in attachment metadata
- Rollback script can reconstruct original state from attachments

## Summary

This design introduces a centralized Rule Library where all rules live independently, with an attachment system that controls which rules apply to which teams.

**Key changes:**

- Rules are no longer owned by teams
- `RuleAttachment` entity links rules to teams with enforcement settings
- Dual approval: library entry + attachment
- Enterprise rules auto-attach to all teams
- Team-specific rules are just library rules attached to one team

**Removed concepts:**

- Team-owned rules
- `Force` flag (replaced by Enterprise layer)
- `TargetTeams`/`TargetUsers` fields
