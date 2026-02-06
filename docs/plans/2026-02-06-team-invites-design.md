# Team Invite System Design

**Date:** 2026-02-06
**Status:** Approved

## Overview

Add ability for users to join teams via invite codes. Enforce one team per user with explicit leave-before-join flow.

## Requirements

- Admins can create expiring invite codes with configurable max uses
- Users join teams by entering invite code
- Users must leave current team before joining another
- Default expiration: 24 hours
- One team per user (already enforced by data model)

## Data Model

### TeamInvite Entity

```go
type TeamInvite struct {
    ID        string    // UUID
    TeamID    string    // FK to teams
    Code      string    // 8-char alphanumeric (e.g., "X7K2M9PQ")
    MaxUses   int       // Configurable limit
    UseCount  int       // Current usage
    ExpiresAt time.Time // Default: created_at + 24h
    CreatedBy string    // User ID of admin who created
    CreatedAt time.Time
}
```

### Database Table

```sql
CREATE TABLE team_invites (
    id UUID PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    code VARCHAR(8) NOT NULL UNIQUE,
    max_uses INT NOT NULL DEFAULT 1,
    use_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMP NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_team_invites_code ON team_invites(code);
CREATE INDEX idx_team_invites_team_id ON team_invites(team_id);
```

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `POST` | `/api/teams/{id}/invites` | Create invite (admin only) | Admin role |
| `GET` | `/api/teams/{id}/invites` | List active invites for team | Admin role |
| `DELETE` | `/api/teams/{id}/invites/{inviteId}` | Revoke an invite | Admin role |
| `POST` | `/api/invites/{code}/join` | Join team using invite code | Authenticated |
| `POST` | `/api/users/me/leave-team` | Leave current team | Authenticated |

### Request/Response Examples

**Create invite:**
```json
POST /api/teams/{id}/invites
{
  "max_uses": 5,
  "expires_in_hours": 24
}

Response 201:
{
  "id": "...",
  "code": "X7K2M9PQ",
  "max_uses": 5,
  "use_count": 0,
  "expires_at": "2026-02-07T12:00:00Z"
}
```

**Join team:**
```json
POST /api/invites/X7K2M9PQ/join

Response 200:
{
  "team_id": "...",
  "team_name": "Test Team"
}
```

### Error Cases

| Scenario | Status | Message |
|----------|--------|---------|
| Join while in team | 409 Conflict | "leave current team first" |
| Invalid/expired code | 404 Not Found | "invite not found or expired" |
| Max uses reached | 410 Gone | "invite has been fully used" |

## Implementation Layers

### Service Layer (`server/services/teams/`)

- `CreateInvite(ctx, teamID, createdBy, maxUses, expiresInHours)` - validate user is admin
- `ListInvites(ctx, teamID)` - return active (non-expired, uses < max) invites
- `RevokeInvite(ctx, teamID, inviteID)` - delete invite
- `JoinByCode(ctx, code, userID)` - atomic: validate, check no team, increment use, update user
- `LeaveTeam(ctx, userID)` - set user's team_id to null

### Database Layer (`server/adapters/postgres/`)

- `team_invite_db.go` - CRUD for invites with proper locking

### Handler Layer (`server/entrypoints/api/handlers/`)

- Extend `teams.go` with invite endpoints
- Add `POST /api/invites/{code}/join` (separate route)
- Add `POST /api/users/me/leave-team` to users handler

### Transaction Safety

`JoinByCode` uses `SELECT FOR UPDATE` on invite row to prevent race conditions on use count.

## Migration Consolidation

Consolidate 23 existing migrations into 7 logical groups:

| # | Name | Contents |
|---|------|----------|
| 1 | `000001_core_tables` | teams, users, rules, projects, agents |
| 2 | `000002_rbac` | permissions, roles, role_permissions, user_roles |
| 3 | `000003_approvals` | approval_configs, rule_approvals |
| 4 | `000004_audit` | audit_entries |
| 5 | `000005_change_management` | change_requests, exception_requests, notifications, channels |
| 6 | `000006_rules_extensions` | enforcement, device_codes, categories, three_file, global |
| 7 | `000007_team_invites` | team_invites (new) |

**Approach:** Clean slate - delete old migrations, create consolidated ones, reset dev databases.

## Seed Data

Add test invite for development:

```sql
INSERT INTO team_invites (id, team_id, code, max_uses, use_count, expires_at, created_by, created_at)
VALUES (
    'e0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',  -- Test Team
    'TESTCODE',
    100,
    0,
    NOW() + INTERVAL '30 days',
    'c0000000-0000-0000-0000-000000000002',  -- admin@test.local
    NOW()
);
```

## Files to Create/Modify

### New Files
- `server/domain/team_invite.go`
- `server/adapters/postgres/team_invite_db.go`
- `server/migrations/000001_core_tables.{up,down}.sql`
- `server/migrations/000002_rbac.{up,down}.sql`
- `server/migrations/000003_approvals.{up,down}.sql`
- `server/migrations/000004_audit.{up,down}.sql`
- `server/migrations/000005_change_management.{up,down}.sql`
- `server/migrations/000006_rules_extensions.{up,down}.sql`
- `server/migrations/000007_team_invites.{up,down}.sql`

### Modified Files
- `server/services/teams/repository.go` - add invite methods
- `server/services/teams/service.go` - add invite + leave logic
- `server/entrypoints/api/handlers/teams.go` - add invite endpoints
- `server/entrypoints/api/handlers/users.go` - add leave-team endpoint
- `tests/infrastructure/seed-data.sql` - add test invite

### Deleted Files
- All existing `server/migrations/0000*.sql` files (23 pairs)
