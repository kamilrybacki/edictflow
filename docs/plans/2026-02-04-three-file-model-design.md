# Three-File Model Design

## Overview

This design transforms Claudeception from managing arbitrary CLAUDE.md files to managing exactly three fixed-location files with rule merging. Rules are created individually in the WebUI and merged into the appropriate CLAUDE.md file based on their level.

## The Three-File Model

| Level | Path | Scope |
|-------|------|-------|
| Enterprise | `/etc/claude-code/CLAUDE.md` | Org-wide policies, applies to all users |
| User | `~/.claude/CLAUDE.md` | Personal standards, admin-pushed + user-created |
| Project | `./CLAUDE.md` | Team-shared instructions, applies to current working directory |

The agent writes project-level rules to `./CLAUDE.md` in whatever directory the developer is working in. All project-level rules apply equally to all projects.

## Rule Model

### Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Rule identifier |
| Content | markdown | The actual instruction for Claude |
| Description | string | Explanation for admins/users |
| Category | foreign key | Reference to category (system or custom) |
| Level | enum | `enterprise`, `user`, `project` |
| Overridable | boolean | Can lower levels contradict this rule? Default: true |
| Priority | integer | Order within category |
| Effective Start | timestamp | When rule becomes active (nullable) |
| Effective End | timestamp | When rule expires (nullable) |
| Target Teams | array | Team IDs this rule applies to (User/Project only) |
| Target Users | array | User IDs this rule applies to (User/Project only) |
| Tags | array | Strings for filtering/organization |

### Categories

- System provides defaults: `Security`, `Coding Standards`, `Testing`, `Documentation`
- Admins can create custom categories per organization
- Users select from available categories when creating rules

## Permissions & Roles

### Role Hierarchy

| Role | Capabilities |
|------|-------------|
| Super Admin | Everything + manage admins |
| Admin | Manage teams, manage all rules at all levels, push rules to users, manage categories |
| User | Create/edit own User-level and Project-level rules only, view rules pushed to them |

### Rule Permissions by Level

| Level | Who can create | Who can edit | Targeting |
|-------|---------------|--------------|-----------|
| Enterprise | Admin only | Admin only | None (applies to all) |
| User | Admin (push to users) + User (own rules) | Creator only | Teams or specific users |
| Project | Admin + User | Creator only | Teams or specific users |

### Team Management

- Only admins can create, rename, or delete teams
- Only admins can add/remove users from teams
- Users can view their team memberships
- Rules can target teams, but users cannot modify team structure

## Rule Merging

### Merge Process

When the agent syncs, it builds each CLAUDE.md file by:

1. Collecting applicable rules for that level (filtered by targeting, effective dates)
2. Grouping rules by category
3. Sorting categories alphabetically (or by admin-defined order)
4. Within each category, sorting rules by priority
5. Rendering into the managed section with level markers

### File Structure

Each managed file has a protected section:

```markdown
# Existing manual content preserved here...

<!-- MANAGED BY CLAUDECEPTION - DO NOT EDIT -->

## Security

[Enterprise] **No Hardcoded Secrets**
Never commit API keys, passwords, or tokens...

[Project] **Input Validation** (overridable)
Validate all user inputs at API boundaries...

## Testing

[Enterprise] **Minimum Coverage**
Maintain 80% test coverage on all new code...

<!-- END CLAUDECEPTION -->

# More manual content can go here...
```

### Override Conflict Handling

- If higher-level rule has `overridable: false` → lower-level contradicting rule is rejected at creation time (WebUI shows error)
- If higher-level rule has `overridable: true` → both appear, Claude sees both and follows the lower-level rule

## Agent Behavior

### File Management

The agent manages a marked section within each file, preserving any manual content outside that section.

If someone manually edits inside the managed section:
- Agent detects the change on next sync
- Restores the managed content to server state
- Sends desktop notification: "Managed CLAUDE.md content restored. Use WebUI to modify rules."

### File Watching

The agent monitors all three paths:
- `/etc/claude-code/CLAUDE.md` (requires elevated permissions to write)
- `~/.claude/CLAUDE.md`
- `./CLAUDE.md` in registered project directories

### Sync Triggers

1. Server push via WebSocket when rules change
2. Agent startup (full sync)
3. Manual `claudeception sync` command
4. File change detected (restore if managed section tampered)

### Offline Behavior

- Agent caches all applicable rules locally (SQLite)
- Uses cached rules to regenerate managed sections
- Effective date filtering uses local clock
- Queues any local rule creations for sync when online

### Project Directory Handling

- Agent writes to `./CLAUDE.md` when user runs `claudeception watch` in a directory
- Watches that directory until `claudeception unwatch`
- Multiple watched directories each get their own `./CLAUDE.md` with same Project-level rules

### File Permissions

- Enterprise file (`/etc/...`) may require sudo; agent prompts or skips with warning
- User file (`~/.claude/...`) created with user permissions
- Project file (`./CLAUDE.md`) created with user permissions

## Database Changes

### Schema Changes

Rules table updates:
- `level` enum: `enterprise`, `user`, `project`
- `category_id` foreign key to categories table
- `overridable` boolean (default true)
- `priority` integer
- `effective_start` timestamp (nullable)
- `effective_end` timestamp (nullable)
- `target_teams` array of team IDs (nullable)
- `target_users` array of user IDs (nullable)
- `tags` array of strings

New categories table:
- `id` - primary key
- `name` - category name
- `is_system` - boolean (system defaults vs admin-created)
- `org_id` - organization reference
- `display_order` - integer for sorting

### Migration

1. Add new columns to rules table
2. Create categories table with system defaults
3. Migrate existing rules to new schema (default to `project` level, assign to generic category)
4. Remove deprecated columns (file paths, etc.)

## API Changes

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/rules` | Create rule (validates level permissions, override conflicts) |
| GET | `/api/rules/merged?level=<level>` | Get merged content for a level |
| GET | `/api/categories` | List available categories |
| POST | `/api/categories` | Create category (admin only) |
| DELETE | `/api/categories/:id` | Delete category (admin only, custom only) |

### WebSocket Messages

- `rules_updated` - Triggers agent sync (includes affected levels)
- `categories_updated` - Refresh category list in agent cache

### Validation Rules

- Enterprise rules: reject if non-admin
- User/Project rules with targeting: reject if user targets teams they don't belong to (unless admin)
- Override conflict: reject if contradicting a `overridable: false` rule at higher level

## WebUI Changes

### Rule Creation Form

| Field | Type | Visibility |
|-------|------|------------|
| Name | text | Always |
| Description | textarea | Always |
| Content | markdown editor | Always |
| Level | dropdown | Always (Enterprise disabled for non-admins) |
| Category | dropdown | Always |
| Overridable | checkbox | Always |
| Priority | number | Always |
| Effective Start | date picker | Always |
| Effective End | date picker | Always |
| Target Teams | multi-select | User/Project levels only |
| Target Users | multi-select | User/Project levels only |
| Tags | tag input | Always |

### Conditional UI

- Level = Enterprise → hide targeting fields, show "Applies to all users"
- Level = User/Project → show targeting fields
- Non-admin users → Enterprise option disabled, can only target self or own teams

### Category Management (Admin Only)

- List system + custom categories
- Create new category with name and display order
- Cannot delete system categories
- Can delete custom categories (with warning if rules use it)

### Rule List View

- Filter by level, category, tags, target team
- Show override status (icon if `overridable: false`)
- Show effective date status (scheduled, active, expired)
- Preview merged output per level

## Implementation Scope

### Server Changes

- `server/domain/` - Update rule and category models
- `server/adapters/postgres/` - New queries for categories, update rule queries
- `server/services/rules/` - Merge logic, override validation, targeting validation
- `server/entrypoints/api/handlers/` - New category handlers, update rule handlers
- `server/entrypoints/api/middleware/` - Enforce admin-only for team/category management

### Agent Changes

- `agent/daemon/` - Track three fixed paths instead of arbitrary files
- `agent/watcher/` - Watch three paths, detect managed section tampering
- `agent/storage/` - Cache rules with new fields, cache categories
- `agent/ws/` - Handle `rules_updated` and `categories_updated` messages
- New merge renderer - Build managed section from cached rules

### Web Changes

- Update rule form with new fields
- Add category management page (admin)
- Conditional field visibility based on level
- Merged preview component
