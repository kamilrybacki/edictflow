# Notification & Change Management System Design

A cross-cutting feature for Claudeception that creates a feedback loop between central control and local autonomy, enabling bidirectional notifications and admin oversight of local changes.

## Overview

**Problem:** Currently, users don't know when their managed CLAUDE.md files change, and admins have no visibility into local modifications.

**Solution:** A notification system that:
1. Alerts users when the agent applies changes to managed files
2. Notifies admins when users modify managed files locally
3. Allows admins to approve, reject, or grant exceptions for local changes
4. Supports configurable enforcement modes per rule

---

## User-Facing Notifications (Agent → User)

### Delivery Methods

- **Desktop notifications** - Native OS notifications (macOS/Windows/Linux) via `gen2brain/beeep` or similar
- **CLI output + logs** - Persistent history via `claudeception changes` command

### Notification Triggers

| Event | Notification |
|-------|--------------|
| Agent applies config update | "CLAUDE.md updated: [rule name]" |
| Change blocked (block mode) | "Change blocked — awaiting admin approval" |
| Change pending (temporary mode) | "Change pending approval (reverts in Xh if not approved)" |
| Change reported (warning mode) | "Change reported to admin" |
| Change approved | "Your change to [file] was approved" |
| Change rejected | "Your change to [file] was rejected" |
| Change auto-reverted | "Change auto-reverted — no approval received" |
| Exception granted | "Exception granted for [file] (expires [date])" |
| Exception denied | "Exception request denied" |

---

## Admin Notifications (User Changes → Admin)

### Delivery Channels

1. **Web UI** - Notifications panel with badge counts, inbox-style interface
2. **Email** - Optional alerts to configured addresses
3. **Webhooks** - Optional integration with Slack/Teams/PagerDuty/etc.

All channels are configurable per team in Settings.

---

## Enforcement Modes

Configurable per rule, determines what happens when a user modifies a managed file.

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Block** (default) | Change immediately reverted, awaits admin approval | Sensitive configs, security rules |
| **Temporary** | Change applies, auto-reverts after timeout if not approved | Standard rules, balanced control |
| **Warning** | Change applies permanently, flagged for visibility | Low-risk customizations |

### Default Behavior

- **Default mode:** Block (strict by default)
- **Default timeout:** 24 hours (for temporary mode)

Both configurable per rule in the rule editor.

---

## Exception System

When an admin rejects a change, users can request an exception.

### Exception Types

| Type | Behavior |
|------|----------|
| **Time-limited** | Exception expires after specified duration |
| **Permanent** | Exception persists until manually revoked |

### Exception Flow

1. Admin rejects a change
2. User runs `claudeception appeal <id>` with justification
3. Admin reviews in Exceptions tab
4. Admin approves (with optional expiry) or denies
5. Agent enforces or skips rule accordingly

---

## Data Model

### New Tables

```sql
-- Change requests from local modifications
CREATE TABLE change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES rules(id),
    agent_id UUID NOT NULL REFERENCES agents(id),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID NOT NULL REFERENCES teams(id),
    file_path TEXT NOT NULL,
    original_hash TEXT NOT NULL,
    modified_hash TEXT NOT NULL,
    diff_content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, rejected, auto_reverted, exception_granted
    enforcement_mode TEXT NOT NULL, -- block, temporary, warning
    timeout_at TIMESTAMPTZ, -- for temporary mode
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

-- Exception requests for rejected changes
CREATE TABLE exception_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id UUID NOT NULL REFERENCES change_requests(id),
    user_id UUID NOT NULL REFERENCES users(id),
    justification TEXT NOT NULL,
    exception_type TEXT NOT NULL, -- time_limited, permanent
    expires_at TIMESTAMPTZ, -- for time_limited
    status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, denied
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

-- Notifications for users
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID REFERENCES teams(id), -- nullable for user-specific
    type TEXT NOT NULL, -- change_detected, approval_required, change_approved, change_rejected, etc.
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Notification channels (email, webhooks)
CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    channel_type TEXT NOT NULL, -- email, webhook
    config JSONB NOT NULL, -- addresses, URL, filters, etc.
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Rule Model Extension

```sql
-- Add to existing rules table
ALTER TABLE rules
ADD COLUMN enforcement_mode TEXT NOT NULL DEFAULT 'block',
ADD COLUMN temporary_timeout_hours INTEGER NOT NULL DEFAULT 24;
```

---

## Agent Architecture

### File Monitoring

Extend existing fsnotify watcher to detect modifications to managed CLAUDE.md files.

```
On file change detected:
├── Compute diff against expected content
├── Look up enforcement_mode for applicable rule
├── If enforcement_mode = "block"
│   ├── Immediately revert file to expected content
│   ├── Send change_detected to server via WebSocket
│   ├── Show desktop notification: "Change blocked, awaiting approval"
│   └── Log to CLI history
├── If enforcement_mode = "temporary"
│   ├── Allow change to persist
│   ├── Send change_detected with timeout_at
│   ├── Show notification: "Change pending approval (reverts in Xh)"
│   └── Start local timer; revert if no approval by timeout
└── If enforcement_mode = "warning"
    ├── Allow change to persist
    ├── Send change_detected (informational)
    └── Show notification: "Change reported to admin"
```

### Local Storage

Store pending changes in SQLite cache for offline resilience:
- Pending change requests (re-sent on reconnect)
- Active exceptions (for local enforcement decisions)
- Notification history

### CLI Additions

| Command | Description |
|---------|-------------|
| `claudeception changes` | List pending/recent change requests and status |
| `claudeception changes [id]` | Show details of a specific change |
| `claudeception appeal [id]` | Submit exception request with justification |

---

## Server Architecture

### New Services

```
services/
├── changes/           # Change request lifecycle
│   ├── service.go     # Create, approve, reject, auto-revert logic
│   └── repository.go  # ChangeRequest CRUD
├── exceptions/        # Exception request handling
│   ├── service.go
│   └── repository.go
├── notifications/     # Notification creation and delivery
│   ├── service.go     # Create notifications, fan-out to channels
│   ├── repository.go  # Notification CRUD
│   └── dispatcher.go  # Background worker for email/webhooks
```

### API Endpoints

**Change Requests:**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/changes` | List change requests (filterable) |
| GET | `/api/changes/:id` | Get change request details with diff |
| POST | `/api/changes/:id/approve` | Approve a pending change |
| POST | `/api/changes/:id/reject` | Reject and trigger revert |

**Exception Requests:**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/exceptions` | List exception requests |
| POST | `/api/exceptions` | Create exception request |
| POST | `/api/exceptions/:id/approve` | Approve exception |
| POST | `/api/exceptions/:id/deny` | Deny exception |

**Notifications:**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/notifications` | List notifications for current user |
| POST | `/api/notifications/:id/read` | Mark as read |
| POST | `/api/notifications/read-all` | Mark all as read |

**Notification Channels:**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/notification-channels` | List team's channels |
| POST | `/api/notification-channels` | Create channel |
| PUT | `/api/notification-channels/:id` | Update channel |
| DELETE | `/api/notification-channels/:id` | Remove channel |

### WebSocket Messages

**Server → Agent:**

| Type | Payload | Description |
|------|---------|-------------|
| `change_approved` | `{ change_id, rule_id }` | Admin approved the change |
| `change_rejected` | `{ change_id, rule_id, revert_to_hash }` | Admin rejected; revert |
| `exception_granted` | `{ change_id, exception_id, expires_at }` | Exception approved |
| `exception_denied` | `{ change_id, exception_id }` | Exception denied |

**Agent → Server:**

| Type | Payload | Description |
|------|---------|-------------|
| `change_detected` | `{ rule_id, file_path, original_hash, modified_hash, diff, enforcement_mode }` | Report modification |
| `change_updated` | `{ change_id, modified_hash, diff }` | Update pending change |
| `exception_request` | `{ change_id, justification, exception_type, requested_duration }` | User appeals |
| `revert_complete` | `{ change_id }` | Confirms file reverted |

### Background Worker

Responsibilities:
- Send emails via SMTP to configured addresses
- POST to webhook URLs with signed JSON payloads
- Check for expired temporary changes (`timeout_at < now()`) and trigger auto-revert
- Retry failed deliveries with exponential backoff (max 5 attempts)

---

## Web UI Components

### Navigation

```
Dashboard (home)
├── ...existing...
├── Changes (new)              # Badge shows pending count
│   ├── Pending
│   ├── History
│   └── Exceptions
├── Notifications (new)        # User's inbox
└── Settings
    └── Notification Channels  # Team config
```

### Changes Page (`/changes`)

- Table: User, Rule, File, Status, Enforcement Mode, Time
- Filters: status, user, rule, date range
- Action buttons: Approve, Reject (for pending)
- Expandable rows with diff viewer

### Change Detail

- Full diff with syntax highlighting
- User/agent info
- Rule info with enforcement mode
- Timeline: detected → resolved
- Actions: Approve, Reject

### Exceptions Tab (`/changes/exceptions`)

- List with justification text
- Filter by status
- Approve (with optional expiry override) or Deny

### Notifications Page (`/notifications`)

- Inbox-style list, grouped by date
- Mark read (individual/bulk)
- Click to navigate to relevant item
- Filter by type, read/unread

### Settings → Channels (`/settings/channels`)

- List existing channels
- Add form: type, config (addresses/URL), event filters
- Test button
- Enable/disable toggle

### Rule Editor Changes

- New "Enforcement" section
- Mode dropdown: Block (default), Allow Temporarily, Allow with Warning
- Timeout input (hours) for temporary mode

---

## Error Handling

| Scenario | Handling |
|----------|----------|
| Agent offline when admin acts | Server queues messages, replays on reconnect |
| Timeout expires (temporary mode) | Worker auto-reverts, notifies user and admins |
| User modifies file while pending | Agent updates existing request with new diff |
| Conflicting exceptions | One active exception per (user, rule, file) tuple |
| Webhook/email failure | Retry with backoff, log errors, don't block workflow |
| Agent cache corrupted | Validate on startup, rebuild from server if invalid |

---

## Flow Examples

### Scenario A: Block Mode Change

1. User modifies `~/projects/foo/CLAUDE.md` (block rule)
2. Agent detects, reverts, sends `change_detected`
3. Agent shows notification: "Change blocked — awaiting approval"
4. Server creates ChangeRequest, notifies admins
5. Admin reviews diff, clicks Approve
6. Server sends `change_approved` to agent
7. Agent re-applies user's change
8. User notified: "Your change was approved"

### Scenario B: Exception Request

1. Admin rejects a change
2. Agent reverts file, user notified
3. User runs `claudeception appeal <id>` with justification
4. Server creates ExceptionRequest, notifies admins
5. Admin approves with 7-day expiry
6. Server sends `exception_granted` to agent
7. Agent re-applies change, stores exception
8. Agent skips enforcement until expiry

---

## Technology Notes

- **Desktop notifications:** `gen2brain/beeep` for cross-platform support
- **Email:** Standard SMTP, configurable in server settings
- **Webhooks:** Signed payloads (HMAC-SHA256) for security
- **Diff generation:** `sergi/go-diff` or similar for unified diffs

---

## Migration Strategy

1. Add new tables (change_requests, exception_requests, notifications, notification_channels)
2. Add enforcement columns to rules table with defaults
3. No data backfill needed — existing rules become "block" mode automatically
