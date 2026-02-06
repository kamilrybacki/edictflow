# Team Invites Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add invite-code-based team joining with expiring codes, configurable max uses, and leave-before-join flow.

**Architecture:** New `TeamInvite` domain entity with dedicated DB adapter. Extend teams handler with invite CRUD endpoints. Add public invite join endpoint and user leave-team endpoint. Consolidate 23 migrations into 7 logical groups.

**Tech Stack:** Go, PostgreSQL, chi router, pgx/v5, bcrypt for code generation

---

## Task 1: Consolidate Migrations - Core Tables

**Files:**
- Delete: `server/migrations/000001_create_teams.{up,down}.sql`
- Delete: `server/migrations/000002_create_users.{up,down}.sql`
- Delete: `server/migrations/000003_create_rules.{up,down}.sql`
- Delete: `server/migrations/000004_create_projects.{up,down}.sql`
- Delete: `server/migrations/000005_create_agents.{up,down}.sql`
- Create: `server/migrations/000001_core_tables.up.sql`
- Create: `server/migrations/000001_core_tables.down.sql`

**Step 1: Delete old migration files (1-5)**

Run:
```bash
rm server/migrations/000001_create_teams.up.sql server/migrations/000001_create_teams.down.sql
rm server/migrations/000002_create_users.up.sql server/migrations/000002_create_users.down.sql
rm server/migrations/000003_create_rules.up.sql server/migrations/000003_create_rules.down.sql
rm server/migrations/000004_create_projects.up.sql server/migrations/000004_create_projects.down.sql
rm server/migrations/000005_create_agents.up.sql server/migrations/000005_create_agents.down.sql
```

**Step 2: Create consolidated up migration**

Create `server/migrations/000001_core_tables.up.sql`:

```sql
-- 000001_core_tables.up.sql
-- Core tables: teams, users, rules, projects, agents

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_teams_name ON teams(name);

-- Users (with all fields from later migrations)
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    avatar_url TEXT,
    auth_provider VARCHAR(50) NOT NULL,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    email_verified BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_created_by ON users(created_by);

-- Rules (with all fields from later migrations)
CREATE TABLE rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    target_layer VARCHAR(50) NOT NULL CHECK (target_layer IN ('enterprise', 'user', 'project')),
    priority_weight INTEGER NOT NULL DEFAULT 0,
    triggers JSONB NOT NULL DEFAULT '[]',
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'approved', 'rejected')),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    submitted_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE,
    enforcement_mode TEXT NOT NULL DEFAULT 'block',
    temporary_timeout_hours INTEGER NOT NULL DEFAULT 24,
    category_id UUID,
    overridable BOOLEAN NOT NULL DEFAULT TRUE,
    effective_start TIMESTAMP WITH TIME ZONE,
    effective_end TIMESTAMP WITH TIME ZONE,
    target_teams UUID[] DEFAULT '{}',
    target_users UUID[] DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    force BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT rules_force_global_only CHECK (force = false OR team_id IS NULL),
    CONSTRAINT rules_global_enterprise_only CHECK (team_id IS NOT NULL OR target_layer = 'enterprise')
);

CREATE INDEX idx_rules_team_id ON rules(team_id);
CREATE INDEX idx_rules_target_layer ON rules(target_layer);
CREATE INDEX idx_rules_status ON rules(status);
CREATE INDEX idx_rules_created_by ON rules(created_by);
CREATE INDEX idx_rules_category_id ON rules(category_id);
CREATE INDEX idx_rules_effective_dates ON rules(effective_start, effective_end);
CREATE INDEX idx_rules_target_teams ON rules USING GIN(target_teams);
CREATE INDEX idx_rules_target_users ON rules USING GIN(target_users);
CREATE INDEX idx_rules_tags ON rules USING GIN(tags);
CREATE INDEX idx_rules_global ON rules(force) WHERE team_id IS NULL;
CREATE INDEX idx_rules_force ON rules(force) WHERE force = true;

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    path_pattern VARCHAR(500) NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_team_id ON projects(team_id);
CREATE INDEX idx_projects_tags ON projects USING GIN(tags);

-- Agents
CREATE TABLE agents (
    id UUID PRIMARY KEY,
    machine_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'offline',
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    cached_config_version INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(machine_id, user_id)
);

CREATE INDEX idx_agents_user_id ON agents(user_id);
CREATE INDEX idx_agents_status ON agents(status);
```

**Step 3: Create consolidated down migration**

Create `server/migrations/000001_core_tables.down.sql`:

```sql
-- 000001_core_tables.down.sql
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS rules;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
```

**Step 4: Commit**

```bash
git add server/migrations/
git commit -m "refactor(migrations): consolidate core tables (1-5) into 000001"
```

---

## Task 2: Consolidate Migrations - RBAC

**Files:**
- Delete: `server/migrations/000006_create_permissions.{up,down}.sql`
- Delete: `server/migrations/000007_create_roles.{up,down}.sql`
- Delete: `server/migrations/000008_create_role_permissions.{up,down}.sql`
- Delete: `server/migrations/000009_modify_users.{up,down}.sql`
- Delete: `server/migrations/000010_create_user_roles.{up,down}.sql`
- Create: `server/migrations/000002_rbac.up.sql`
- Create: `server/migrations/000002_rbac.down.sql`

**Step 1: Delete old migration files (6-10)**

Run:
```bash
rm server/migrations/000006_create_permissions.up.sql server/migrations/000006_create_permissions.down.sql
rm server/migrations/000007_create_roles.up.sql server/migrations/000007_create_roles.down.sql
rm server/migrations/000008_create_role_permissions.up.sql server/migrations/000008_create_role_permissions.down.sql
rm server/migrations/000009_modify_users.up.sql server/migrations/000009_modify_users.down.sql
rm server/migrations/000010_create_user_roles.up.sql server/migrations/000010_create_user_roles.down.sql
```

**Step 2: Create consolidated up migration**

Create `server/migrations/000002_rbac.up.sql`:

```sql
-- 000002_rbac.up.sql
-- RBAC: permissions, roles, role_permissions, user_roles

-- Permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed built-in permissions
INSERT INTO permissions (id, code, description, category) VALUES
    ('a0000001-0000-0000-0000-000000000001', 'create_rules', 'Create new rules as drafts', 'rules'),
    ('a0000001-0000-0000-0000-000000000002', 'edit_rules', 'Edit existing rules', 'rules'),
    ('a0000001-0000-0000-0000-000000000003', 'delete_rules', 'Delete rules', 'rules'),
    ('a0000001-0000-0000-0000-000000000004', 'approve_local', 'Approve local-scoped rules', 'rules'),
    ('a0000001-0000-0000-0000-000000000005', 'approve_project', 'Approve project-scoped rules', 'rules'),
    ('a0000001-0000-0000-0000-000000000006', 'approve_global', 'Approve global-scoped rules', 'rules'),
    ('a0000001-0000-0000-0000-000000000007', 'approve_enterprise', 'Approve enterprise-scoped rules', 'rules'),
    ('a0000001-0000-0000-0000-000000000008', 'manage_users', 'Create, edit, and deactivate users', 'users'),
    ('a0000001-0000-0000-0000-000000000009', 'manage_roles', 'Create and edit roles and permissions', 'admin'),
    ('a0000001-0000-0000-0000-00000000000a', 'manage_team_settings', 'Edit team configuration', 'teams'),
    ('a0000001-0000-0000-0000-00000000000b', 'view_audit_log', 'View change history', 'admin'),
    ('a0000001-0000-0000-0000-00000000000c', 'manage_team_invites', 'Create and revoke team invites', 'teams');

-- Roles
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    hierarchy_level INT NOT NULL,
    parent_role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, team_id)
);

CREATE INDEX idx_roles_team ON roles(team_id);
CREATE INDEX idx_roles_parent ON roles(parent_role_id);

-- Seed system roles (global, team_id = NULL)
INSERT INTO roles (id, name, description, hierarchy_level, is_system) VALUES
    ('b0000001-0000-0000-0000-000000000001', 'Member', 'Default role for new users', 1, true),
    ('b0000001-0000-0000-0000-000000000002', 'Admin', 'Full system access', 100, true);

-- Role permissions junction
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- Assign permissions to Member role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('b0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001'),
    ('b0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000004');

-- Assign all permissions to Admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000001-0000-0000-0000-000000000002', id FROM permissions;

-- User roles junction
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
```

**Step 3: Create consolidated down migration**

Create `server/migrations/000002_rbac.down.sql`:

```sql
-- 000002_rbac.down.sql
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
```

**Step 4: Commit**

```bash
git add server/migrations/
git commit -m "refactor(migrations): consolidate RBAC (6-10) into 000002"
```

---

## Task 3: Consolidate Migrations - Approvals

**Files:**
- Delete: `server/migrations/000011_create_approval_configs.{up,down}.sql`
- Delete: `server/migrations/000012_create_rule_approvals.{up,down}.sql`
- Delete: `server/migrations/000013_modify_rules.{up,down}.sql`
- Create: `server/migrations/000003_approvals.up.sql`
- Create: `server/migrations/000003_approvals.down.sql`

**Step 1: Delete old migration files (11-13)**

Run:
```bash
rm server/migrations/000011_create_approval_configs.up.sql server/migrations/000011_create_approval_configs.down.sql
rm server/migrations/000012_create_rule_approvals.up.sql server/migrations/000012_create_rule_approvals.down.sql
rm server/migrations/000013_modify_rules.up.sql server/migrations/000013_modify_rules.down.sql
```

**Step 2: Create consolidated up migration**

Create `server/migrations/000003_approvals.up.sql`:

```sql
-- 000003_approvals.up.sql
-- Approvals: approval_configs, rule_approvals

-- Approval configs
CREATE TABLE approval_configs (
    id UUID PRIMARY KEY,
    scope VARCHAR(50) NOT NULL,
    required_permission VARCHAR(100) NOT NULL,
    required_count INT NOT NULL DEFAULT 1,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(scope, team_id)
);

CREATE INDEX idx_approval_configs_scope ON approval_configs(scope);
CREATE INDEX idx_approval_configs_team ON approval_configs(team_id);

-- Seed global defaults (team_id = NULL)
INSERT INTO approval_configs (id, scope, required_permission, required_count) VALUES
    ('c0000001-0000-0000-0000-000000000001', 'local', 'approve_local', 1),
    ('c0000001-0000-0000-0000-000000000002', 'project', 'approve_project', 1),
    ('c0000001-0000-0000-0000-000000000003', 'global', 'approve_global', 2),
    ('c0000001-0000-0000-0000-000000000004', 'enterprise', 'approve_enterprise', 3);

-- Rule approvals
CREATE TABLE rule_approvals (
    id UUID PRIMARY KEY,
    rule_id UUID REFERENCES rules(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('approved', 'rejected')),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rule_approvals_rule ON rule_approvals(rule_id);
CREATE INDEX idx_rule_approvals_user ON rule_approvals(user_id);
```

**Step 3: Create consolidated down migration**

Create `server/migrations/000003_approvals.down.sql`:

```sql
-- 000003_approvals.down.sql
DROP TABLE IF EXISTS rule_approvals;
DROP TABLE IF EXISTS approval_configs;
```

**Step 4: Commit**

```bash
git add server/migrations/
git commit -m "refactor(migrations): consolidate approvals (11-13) into 000003"
```

---

## Task 4: Consolidate Migrations - Audit

**Files:**
- Delete: `server/migrations/000014_create_audit_entries.{up,down}.sql`
- Create: `server/migrations/000004_audit.up.sql`
- Create: `server/migrations/000004_audit.down.sql`

**Step 1: Delete old migration file (14)**

Run:
```bash
rm server/migrations/000014_create_audit_entries.up.sql server/migrations/000014_create_audit_entries.down.sql
```

**Step 2: Create consolidated up migration**

Create `server/migrations/000004_audit.up.sql`:

```sql
-- 000004_audit.up.sql
-- Audit entries

CREATE TABLE audit_entries (
    id UUID PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    changes JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_entity ON audit_entries(entity_type, entity_id);
CREATE INDEX idx_audit_actor ON audit_entries(actor_id);
CREATE INDEX idx_audit_created ON audit_entries(created_at);
CREATE INDEX idx_audit_action ON audit_entries(action);
```

**Step 3: Create consolidated down migration**

Create `server/migrations/000004_audit.down.sql`:

```sql
-- 000004_audit.down.sql
DROP TABLE IF EXISTS audit_entries;
```

**Step 4: Commit**

```bash
git add server/migrations/
git commit -m "refactor(migrations): consolidate audit (14) into 000004"
```

---

## Task 5: Consolidate Migrations - Change Management

**Files:**
- Delete: `server/migrations/000015_add_change_requests.{up,down}.sql`
- Delete: `server/migrations/000016_add_exception_requests.{up,down}.sql`
- Delete: `server/migrations/000017_add_notifications.{up,down}.sql`
- Delete: `server/migrations/000018_add_notification_channels.{up,down}.sql`
- Create: `server/migrations/000005_change_management.up.sql`
- Create: `server/migrations/000005_change_management.down.sql`

**Step 1: Delete old migration files (15-18)**

Run:
```bash
rm server/migrations/000015_add_change_requests.up.sql server/migrations/000015_add_change_requests.down.sql
rm server/migrations/000016_add_exception_requests.up.sql server/migrations/000016_add_exception_requests.down.sql
rm server/migrations/000017_add_notifications.up.sql server/migrations/000017_add_notifications.down.sql
rm server/migrations/000018_add_notification_channels.up.sql server/migrations/000018_add_notification_channels.down.sql
```

**Step 2: Create consolidated up migration**

Create `server/migrations/000005_change_management.up.sql`:

```sql
-- 000005_change_management.up.sql
-- Change management: change_requests, exception_requests, notifications, notification_channels

-- Change requests
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
    status TEXT NOT NULL DEFAULT 'pending',
    enforcement_mode TEXT NOT NULL,
    timeout_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

CREATE INDEX idx_change_requests_team_status ON change_requests(team_id, status);
CREATE INDEX idx_change_requests_rule_id ON change_requests(rule_id);
CREATE INDEX idx_change_requests_agent_id ON change_requests(agent_id);
CREATE INDEX idx_change_requests_timeout_at ON change_requests(timeout_at);

-- Exception requests
CREATE TABLE exception_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id UUID NOT NULL REFERENCES change_requests(id),
    user_id UUID NOT NULL REFERENCES users(id),
    justification TEXT NOT NULL,
    exception_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

CREATE INDEX idx_exception_requests_change_request_id ON exception_requests(change_request_id);
CREATE INDEX idx_exception_requests_status ON exception_requests(status);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID REFERENCES teams(id),
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_read_at ON notifications(user_id, read_at);
CREATE INDEX idx_notifications_team_id ON notifications(team_id);

-- Notification channels
CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    channel_type TEXT NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_channels_team_enabled ON notification_channels(team_id, enabled);
```

**Step 3: Create consolidated down migration**

Create `server/migrations/000005_change_management.down.sql`:

```sql
-- 000005_change_management.down.sql
DROP TABLE IF EXISTS notification_channels;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS exception_requests;
DROP TABLE IF EXISTS change_requests;
```

**Step 4: Commit**

```bash
git add server/migrations/
git commit -m "refactor(migrations): consolidate change management (15-18) into 000005"
```

---

## Task 6: Consolidate Migrations - Extensions

**Files:**
- Delete: `server/migrations/000019_add_rule_enforcement.{up,down}.sql`
- Delete: `server/migrations/000020_add_device_codes.{up,down}.sql`
- Delete: `server/migrations/000021_create_categories.{up,down}.sql`
- Delete: `server/migrations/000022_extend_rules_three_file.{up,down}.sql`
- Delete: `server/migrations/000023_add_global_rules.{up,down}.sql`
- Create: `server/migrations/000006_extensions.up.sql`
- Create: `server/migrations/000006_extensions.down.sql`

**Step 1: Delete old migration files (19-23)**

Run:
```bash
rm server/migrations/000019_add_rule_enforcement.up.sql server/migrations/000019_add_rule_enforcement.down.sql
rm server/migrations/000020_add_device_codes.up.sql server/migrations/000020_add_device_codes.down.sql
rm server/migrations/000021_create_categories.up.sql server/migrations/000021_create_categories.down.sql
rm server/migrations/000022_extend_rules_three_file.up.sql server/migrations/000022_extend_rules_three_file.down.sql
rm server/migrations/000023_add_global_rules.up.sql server/migrations/000023_add_global_rules.down.sql
```

**Step 2: Create consolidated up migration**

Create `server/migrations/000006_extensions.up.sql`:

```sql
-- 000006_extensions.up.sql
-- Extensions: device_codes, categories (rules extensions already in 000001)

-- Device codes for CLI auth
CREATE TABLE device_codes (
    device_code TEXT PRIMARY KEY,
    user_code TEXT UNIQUE NOT NULL,
    user_id UUID REFERENCES users(id),
    client_id TEXT NOT NULL DEFAULT 'edictflow-cli',
    expires_at TIMESTAMPTZ NOT NULL,
    authorized_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_device_codes_user_code ON device_codes(user_code);
CREATE INDEX idx_device_codes_expires_at ON device_codes(expires_at);

-- Categories for rules
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    org_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_categories_name_org ON categories(name, COALESCE(org_id, '00000000-0000-0000-0000-000000000000'));

-- Insert system default categories
INSERT INTO categories (id, name, is_system, org_id, display_order) VALUES
    (gen_random_uuid(), 'Security', TRUE, NULL, 1),
    (gen_random_uuid(), 'Coding Standards', TRUE, NULL, 2),
    (gen_random_uuid(), 'Testing', TRUE, NULL, 3),
    (gen_random_uuid(), 'Documentation', TRUE, NULL, 4);

-- Add foreign key from rules to categories (rules table already has category_id column)
ALTER TABLE rules ADD CONSTRAINT fk_rules_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
```

**Step 3: Create consolidated down migration**

Create `server/migrations/000006_extensions.down.sql`:

```sql
-- 000006_extensions.down.sql
ALTER TABLE rules DROP CONSTRAINT IF EXISTS fk_rules_category;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS device_codes;
```

**Step 4: Commit**

```bash
git add server/migrations/
git commit -m "refactor(migrations): consolidate extensions (19-23) into 000006"
```

---

## Task 7: Create Team Invites Migration

**Files:**
- Create: `server/migrations/000007_team_invites.up.sql`
- Create: `server/migrations/000007_team_invites.down.sql`

**Step 1: Create up migration**

Create `server/migrations/000007_team_invites.up.sql`:

```sql
-- 000007_team_invites.up.sql
-- Team invites for invite-code-based joining

CREATE TABLE team_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    code VARCHAR(8) NOT NULL UNIQUE,
    max_uses INT NOT NULL DEFAULT 1,
    use_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_team_invites_code ON team_invites(code);
CREATE INDEX idx_team_invites_team_id ON team_invites(team_id);
CREATE INDEX idx_team_invites_expires_at ON team_invites(expires_at);
```

**Step 2: Create down migration**

Create `server/migrations/000007_team_invites.down.sql`:

```sql
-- 000007_team_invites.down.sql
DROP TABLE IF EXISTS team_invites;
```

**Step 3: Commit**

```bash
git add server/migrations/
git commit -m "feat(migrations): add team_invites table"
```

---

## Task 8: Create TeamInvite Domain Entity

**Files:**
- Create: `server/domain/team_invite.go`
- Test: `server/domain/team_invite_test.go`

**Step 1: Write the test**

Create `server/domain/team_invite_test.go`:

```go
package domain

import (
	"testing"
	"time"
)

func TestNewTeamInvite(t *testing.T) {
	teamID := "team-123"
	createdBy := "user-456"
	maxUses := 5
	expiresInHours := 24

	invite := NewTeamInvite(teamID, createdBy, maxUses, expiresInHours)

	if invite.TeamID != teamID {
		t.Errorf("expected TeamID %s, got %s", teamID, invite.TeamID)
	}
	if invite.CreatedBy != createdBy {
		t.Errorf("expected CreatedBy %s, got %s", createdBy, invite.CreatedBy)
	}
	if invite.MaxUses != maxUses {
		t.Errorf("expected MaxUses %d, got %d", maxUses, invite.MaxUses)
	}
	if invite.UseCount != 0 {
		t.Errorf("expected UseCount 0, got %d", invite.UseCount)
	}
	if len(invite.Code) != 8 {
		t.Errorf("expected Code length 8, got %d", len(invite.Code))
	}
	if invite.ID == "" {
		t.Error("expected ID to be set")
	}

	expectedExpiry := time.Now().Add(time.Duration(expiresInHours) * time.Hour)
	if invite.ExpiresAt.Before(expectedExpiry.Add(-time.Minute)) || invite.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("expected ExpiresAt around %v, got %v", expectedExpiry, invite.ExpiresAt)
	}
}

func TestTeamInvite_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		invite   TeamInvite
		expected bool
	}{
		{
			name: "valid invite",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  2,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: true,
		},
		{
			name: "expired invite",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  2,
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expected: false,
		},
		{
			name: "max uses reached",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  5,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "max uses exceeded",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  6,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.invite.IsValid()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTeamInvite_IncrementUseCount(t *testing.T) {
	invite := TeamInvite{UseCount: 2}
	invite.IncrementUseCount()
	if invite.UseCount != 3 {
		t.Errorf("expected UseCount 3, got %d", invite.UseCount)
	}
}

func TestGenerateInviteCode(t *testing.T) {
	code := GenerateInviteCode()

	if len(code) != 8 {
		t.Errorf("expected code length 8, got %d", len(code))
	}

	// Check all characters are alphanumeric uppercase
	for _, c := range code {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Errorf("unexpected character in code: %c", c)
		}
	}

	// Check uniqueness (generate 100 codes, all should be different)
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		c := GenerateInviteCode()
		if codes[c] {
			t.Errorf("duplicate code generated: %s", c)
		}
		codes[c] = true
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain -run TestTeamInvite -v`

Expected: FAIL with "undefined: NewTeamInvite"

**Step 3: Write minimal implementation**

Create `server/domain/team_invite.go`:

```go
package domain

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type TeamInvite struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	Code      string    `json:"code"`
	MaxUses   int       `json:"max_uses"`
	UseCount  int       `json:"use_count"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

const inviteCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excludes I, O, 0, 1 for readability

func GenerateInviteCode() string {
	code := make([]byte, 8)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeChars))))
		code[i] = inviteCodeChars[n.Int64()]
	}
	return string(code)
}

func NewTeamInvite(teamID, createdBy string, maxUses, expiresInHours int) TeamInvite {
	return TeamInvite{
		ID:        uuid.New().String(),
		TeamID:    teamID,
		Code:      GenerateInviteCode(),
		MaxUses:   maxUses,
		UseCount:  0,
		ExpiresAt: time.Now().Add(time.Duration(expiresInHours) * time.Hour),
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}
}

func (i *TeamInvite) IsValid() bool {
	return i.UseCount < i.MaxUses && time.Now().Before(i.ExpiresAt)
}

func (i *TeamInvite) IncrementUseCount() {
	i.UseCount++
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain -run TestTeamInvite -v`

Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/team_invite.go server/domain/team_invite_test.go
git commit -m "feat(domain): add TeamInvite entity with code generation"
```

---

## Task 9: Create TeamInvite Database Adapter

**Files:**
- Create: `server/adapters/postgres/team_invite_db.go`
- Test: `server/adapters/postgres/team_invite_db_test.go`

**Step 1: Write the database adapter**

Create `server/adapters/postgres/team_invite_db.go`:

```go
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrInviteNotFound = errors.New("invite not found")
	ErrInviteExpired  = errors.New("invite expired or max uses reached")
)

type TeamInviteDB struct {
	pool *pgxpool.Pool
}

func NewTeamInviteDB(pool *pgxpool.Pool) *TeamInviteDB {
	return &TeamInviteDB{pool: pool}
}

func (db *TeamInviteDB) Create(ctx context.Context, invite domain.TeamInvite) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO team_invites (id, team_id, code, max_uses, use_count, expires_at, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, invite.ID, invite.TeamID, invite.Code, invite.MaxUses, invite.UseCount, invite.ExpiresAt, invite.CreatedBy, invite.CreatedAt)
	return err
}

func (db *TeamInviteDB) GetByCode(ctx context.Context, code string) (domain.TeamInvite, error) {
	var invite domain.TeamInvite
	err := db.pool.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE code = $1
	`, code).Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamInvite{}, ErrInviteNotFound
	}
	return invite, err
}

func (db *TeamInviteDB) GetByID(ctx context.Context, id string) (domain.TeamInvite, error) {
	var invite domain.TeamInvite
	err := db.pool.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE id = $1
	`, id).Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamInvite{}, ErrInviteNotFound
	}
	return invite, err
}

func (db *TeamInviteDB) ListByTeam(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE team_id = $1 AND expires_at > NOW() AND use_count < max_uses
		ORDER BY created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []domain.TeamInvite
	for rows.Next() {
		var invite domain.TeamInvite
		if err := rows.Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, rows.Err()
}

func (db *TeamInviteDB) Delete(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `DELETE FROM team_invites WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInviteNotFound
	}
	return nil
}

// IncrementUseCountAtomic atomically increments use_count and returns the updated invite.
// Uses SELECT FOR UPDATE to prevent race conditions.
// Returns ErrInviteExpired if the invite is expired or max uses reached.
func (db *TeamInviteDB) IncrementUseCountAtomic(ctx context.Context, code string) (domain.TeamInvite, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return domain.TeamInvite{}, err
	}
	defer tx.Rollback(ctx)

	var invite domain.TeamInvite
	err = tx.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE code = $1
		FOR UPDATE
	`, code).Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamInvite{}, ErrInviteNotFound
	}
	if err != nil {
		return domain.TeamInvite{}, err
	}

	if !invite.IsValid() {
		return domain.TeamInvite{}, ErrInviteExpired
	}

	_, err = tx.Exec(ctx, `
		UPDATE team_invites SET use_count = use_count + 1 WHERE id = $1
	`, invite.ID)
	if err != nil {
		return domain.TeamInvite{}, err
	}

	invite.UseCount++

	if err := tx.Commit(ctx); err != nil {
		return domain.TeamInvite{}, err
	}

	return invite, nil
}
```

**Step 2: Commit**

```bash
git add server/adapters/postgres/team_invite_db.go
git commit -m "feat(adapters): add TeamInviteDB with atomic use count increment"
```

---

## Task 10: Extend Teams Repository with Invite Methods

**Files:**
- Modify: `server/services/teams/repository.go`

**Step 1: Add invite interface and methods**

Edit `server/services/teams/repository.go` - add after the existing `DB` interface:

```go
type InviteDB interface {
	Create(ctx context.Context, invite domain.TeamInvite) error
	GetByCode(ctx context.Context, code string) (domain.TeamInvite, error)
	GetByID(ctx context.Context, id string) (domain.TeamInvite, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.TeamInvite, error)
	Delete(ctx context.Context, id string) error
	IncrementUseCountAtomic(ctx context.Context, code string) (domain.TeamInvite, error)
}
```

Update `Repository` struct:

```go
type Repository struct {
	db       DB
	inviteDB InviteDB
}

func NewRepository(db DB, inviteDB InviteDB) *Repository {
	return &Repository{db: db, inviteDB: inviteDB}
}
```

Add invite methods:

```go
func (r *Repository) CreateInvite(ctx context.Context, invite domain.TeamInvite) error {
	return r.inviteDB.Create(ctx, invite)
}

func (r *Repository) GetInviteByCode(ctx context.Context, code string) (domain.TeamInvite, error) {
	return r.inviteDB.GetByCode(ctx, code)
}

func (r *Repository) GetInviteByID(ctx context.Context, id string) (domain.TeamInvite, error) {
	return r.inviteDB.GetByID(ctx, id)
}

func (r *Repository) ListInvites(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	return r.inviteDB.ListByTeam(ctx, teamID)
}

func (r *Repository) DeleteInvite(ctx context.Context, id string) error {
	return r.inviteDB.Delete(ctx, id)
}

func (r *Repository) UseInvite(ctx context.Context, code string) (domain.TeamInvite, error) {
	return r.inviteDB.IncrementUseCountAtomic(ctx, code)
}
```

**Step 2: Update repository tests if they exist**

Check `server/services/teams/repository_test.go` and update mock if needed.

**Step 3: Commit**

```bash
git add server/services/teams/repository.go
git commit -m "feat(teams): extend repository with invite methods"
```

---

## Task 11: Add User Leave Team to User Service

**Files:**
- Modify: `server/services/users/service.go`

**Step 1: Add LeaveTeam method**

Edit `server/services/users/service.go` - add new error and method:

```go
var ErrNotInTeam = errors.New("user is not in a team")

func (s *Service) LeaveTeam(ctx context.Context, userID string) error {
	user, err := s.userDB.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.TeamID == nil {
		return ErrNotInTeam
	}

	user.TeamID = nil
	return s.userDB.Update(ctx, user)
}

func (s *Service) JoinTeam(ctx context.Context, userID, teamID string) error {
	user, err := s.userDB.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.TeamID != nil {
		return errors.New("user already in a team")
	}

	user.TeamID = &teamID
	return s.userDB.Update(ctx, user)
}
```

**Step 2: Commit**

```bash
git add server/services/users/service.go
git commit -m "feat(users): add LeaveTeam and JoinTeam methods"
```

---

## Task 12: Add Invite Handlers to Teams Handler

**Files:**
- Modify: `server/entrypoints/api/handlers/teams.go`

**Step 1: Add invite types and handler methods**

Edit `server/entrypoints/api/handlers/teams.go` - add after existing types:

```go
type CreateInviteRequest struct {
	MaxUses        int `json:"max_uses"`
	ExpiresInHours int `json:"expires_in_hours,omitempty"`
}

type InviteResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	MaxUses   int    `json:"max_uses"`
	UseCount  int    `json:"use_count"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

func inviteToResponse(invite domain.TeamInvite) InviteResponse {
	return InviteResponse{
		ID:        invite.ID,
		Code:      invite.Code,
		MaxUses:   invite.MaxUses,
		UseCount:  invite.UseCount,
		ExpiresAt: invite.ExpiresAt.Format(time.RFC3339),
		CreatedAt: invite.CreatedAt.Format(time.RFC3339),
	}
}
```

Update `TeamService` interface:

```go
type TeamService interface {
	Create(ctx context.Context, name string) (domain.Team, error)
	GetByID(ctx context.Context, id string) (domain.Team, error)
	List(ctx context.Context) ([]domain.Team, error)
	Update(ctx context.Context, team domain.Team) error
	Delete(ctx context.Context, id string) error
	// Invite methods
	CreateInvite(ctx context.Context, teamID, createdBy string, maxUses, expiresInHours int) (domain.TeamInvite, error)
	ListInvites(ctx context.Context, teamID string) ([]domain.TeamInvite, error)
	DeleteInvite(ctx context.Context, teamID, inviteID string) error
}
```

Add handler methods:

```go
func (h *TeamsHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}
	if req.ExpiresInHours <= 0 {
		req.ExpiresInHours = 24
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	invite, err := h.service.CreateInvite(r.Context(), teamID, userID, req.MaxUses, req.ExpiresInHours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inviteToResponse(invite))
}

func (h *TeamsHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	invites, err := h.service.ListInvites(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []InviteResponse
	for _, invite := range invites {
		response = append(response, inviteToResponse(invite))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TeamsHandler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")
	inviteID := chi.URLParam(r, "inviteId")

	if err := h.service.DeleteInvite(r.Context(), teamID, inviteID); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "invite not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

Update `RegisterRoutes`:

```go
func (h *TeamsHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}/settings", h.UpdateSettings)
	r.Delete("/{id}", h.Delete)

	// Invite routes
	r.Post("/{id}/invites", h.CreateInvite)
	r.Get("/{id}/invites", h.ListInvites)
	r.Delete("/{id}/invites/{inviteId}", h.DeleteInvite)
}
```

**Step 2: Commit**

```bash
git add server/entrypoints/api/handlers/teams.go
git commit -m "feat(handlers): add team invite CRUD endpoints"
```

---

## Task 13: Add Join Invite Handler

**Files:**
- Create: `server/entrypoints/api/handlers/invites.go`

**Step 1: Create invite join handler**

Create `server/entrypoints/api/handlers/invites.go`:

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type InviteService interface {
	JoinByCode(ctx context.Context, code, userID string) (domain.Team, error)
}

type InvitesHandler struct {
	service InviteService
}

func NewInvitesHandler(service InviteService) *InvitesHandler {
	return &InvitesHandler{service: service}
}

type JoinTeamResponse struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
}

func (h *InvitesHandler) Join(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "invite code required", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	team, err := h.service.JoinByCode(r.Context(), code, userID)
	if err != nil {
		switch err.Error() {
		case "invite not found", "invite expired or max uses reached":
			http.Error(w, "invite not found or expired", http.StatusNotFound)
		case "user already in a team":
			http.Error(w, "leave current team first", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JoinTeamResponse{
		TeamID:   team.ID,
		TeamName: team.Name,
	})
}

func (h *InvitesHandler) RegisterRoutes(r chi.Router) {
	r.Post("/{code}/join", h.Join)
}
```

**Step 2: Commit**

```bash
git add server/entrypoints/api/handlers/invites.go
git commit -m "feat(handlers): add invite join endpoint"
```

---

## Task 14: Add Leave Team Endpoint to Users Handler

**Files:**
- Modify: `server/entrypoints/api/handlers/users.go`

**Step 1: Update UsersService interface and add handler**

Edit `server/entrypoints/api/handlers/users.go` - update interface:

```go
type UsersService interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error)
	Update(ctx context.Context, user domain.User) error
	Deactivate(ctx context.Context, id string) error
	GetWithRolesAndPermissions(ctx context.Context, id string) (domain.User, error)
	LeaveTeam(ctx context.Context, userID string) error
}
```

Add handler method:

```go
func (h *UsersHandler) LeaveTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.LeaveTeam(r.Context(), userID); err != nil {
		if err.Error() == "user is not in a team" {
			http.Error(w, "not in a team", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

Update `RegisterRoutes`:

```go
func (h *UsersHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Deactivate)
	r.Post("/me/leave-team", h.LeaveTeam)
}
```

**Step 2: Commit**

```bash
git add server/entrypoints/api/handlers/users.go
git commit -m "feat(handlers): add leave-team endpoint"
```

---

## Task 15: Create Teams Service with Invite Logic

**Files:**
- Create: `server/services/teams/service.go`

**Step 1: Create service**

Create `server/services/teams/service.go`:

```go
package teams

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrUserAlreadyInTeam = errors.New("user already in a team")
	ErrInviteNotFound    = errors.New("invite not found")
	ErrInviteExpired     = errors.New("invite expired or max uses reached")
)

type UserDB interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
}

type Service struct {
	repo   *Repository
	userDB UserDB
}

func NewService(repo *Repository, userDB UserDB) *Service {
	return &Service{repo: repo, userDB: userDB}
}

func (s *Service) Create(ctx context.Context, name string) (domain.Team, error) {
	team := domain.NewTeam(name)
	if err := team.Validate(); err != nil {
		return domain.Team{}, err
	}
	if err := s.repo.Create(ctx, team); err != nil {
		return domain.Team{}, err
	}
	return team, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Team, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Team, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, team domain.Team) error {
	return s.repo.Update(ctx, team)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Invite methods

func (s *Service) CreateInvite(ctx context.Context, teamID, createdBy string, maxUses, expiresInHours int) (domain.TeamInvite, error) {
	// Verify team exists
	if _, err := s.repo.GetByID(ctx, teamID); err != nil {
		return domain.TeamInvite{}, err
	}

	invite := domain.NewTeamInvite(teamID, createdBy, maxUses, expiresInHours)
	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return domain.TeamInvite{}, err
	}
	return invite, nil
}

func (s *Service) ListInvites(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	return s.repo.ListInvites(ctx, teamID)
}

func (s *Service) DeleteInvite(ctx context.Context, teamID, inviteID string) error {
	// Verify invite belongs to team
	invite, err := s.repo.GetInviteByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite.TeamID != teamID {
		return ErrInviteNotFound
	}
	return s.repo.DeleteInvite(ctx, inviteID)
}

func (s *Service) JoinByCode(ctx context.Context, code, userID string) (domain.Team, error) {
	// Get user and check not already in team
	user, err := s.userDB.GetByID(ctx, userID)
	if err != nil {
		return domain.Team{}, err
	}
	if user.TeamID != nil {
		return domain.Team{}, ErrUserAlreadyInTeam
	}

	// Use invite (atomic increment)
	invite, err := s.repo.UseInvite(ctx, code)
	if err != nil {
		return domain.Team{}, err
	}

	// Get team
	team, err := s.repo.GetByID(ctx, invite.TeamID)
	if err != nil {
		return domain.Team{}, err
	}

	// Update user's team
	user.TeamID = &team.ID
	if err := s.userDB.Update(ctx, user); err != nil {
		return domain.Team{}, err
	}

	return team, nil
}
```

**Step 2: Commit**

```bash
git add server/services/teams/service.go
git commit -m "feat(teams): add service with invite logic"
```

---

## Task 16: Wire Up Router with Invite Endpoints

**Files:**
- Modify: `server/entrypoints/api/router.go`

**Step 1: Add InviteService to Config**

Edit `server/entrypoints/api/router.go` - add to Config struct:

```go
InviteService handlers.InviteService
```

**Step 2: Add invite routes**

Add after teams route in `NewRouter`:

```go
// Invite join route (authenticated, but not team-specific)
if cfg.InviteService != nil {
	r.Route("/invites", func(r chi.Router) {
		h := handlers.NewInvitesHandler(cfg.InviteService)
		h.RegisterRoutes(r)
	})
}
```

**Step 3: Add users routes (if not already present)**

Add users route in `/api/v1`:

```go
if cfg.UserService != nil {
	r.Route("/users", func(r chi.Router) {
		h := handlers.NewUsersHandler(cfg.UserService)
		h.RegisterRoutes(r)
	})
}
```

**Step 4: Commit**

```bash
git add server/entrypoints/api/router.go
git commit -m "feat(router): wire up invite and user endpoints"
```

---

## Task 17: Update Seed Data with Test Invite

**Files:**
- Modify: `tests/infrastructure/seed-data.sql`

**Step 1: Add test invite**

Edit `tests/infrastructure/seed-data.sql` - add at end:

```sql
-- ============================================
-- Test Team Invite
-- ============================================

INSERT INTO team_invites (id, team_id, code, max_uses, use_count, expires_at, created_by, created_at)
VALUES (
    'e0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'TESTCODE',
    100,
    0,
    NOW() + INTERVAL '30 days',
    'c0000000-0000-0000-0000-000000000002',
    NOW()
) ON CONFLICT (id) DO NOTHING;
```

**Step 2: Commit**

```bash
git add tests/infrastructure/seed-data.sql
git commit -m "feat(seed): add test invite code TESTCODE"
```

---

## Task 18: Update Main Server Initialization

**Files:**
- Modify: `server/cmd/server/main.go`

**Step 1: Initialize invite DB and service**

Add imports and initialization for TeamInviteDB and wire up the service.

Add after `teamDB` initialization:

```go
teamInviteDB := postgres.NewTeamInviteDB(pool)
```

Update team service creation to use the new Service:

```go
teamRepo := teams.NewRepository(teamDB, teamInviteDB)
teamService := teams.NewService(teamRepo, userDB)
```

Update router config:

```go
InviteService: teamService,
```

**Step 2: Commit**

```bash
git add server/cmd/server/main.go
git commit -m "feat(main): wire up team invites in server initialization"
```

---

## Task 19: Reset Dev Database and Test

**Step 1: Stop and remove volumes**

Run:
```bash
docker-compose down -v
```

**Step 2: Start fresh**

Run:
```bash
docker-compose up -d
```

**Step 3: Verify migrations applied**

Run:
```bash
docker-compose logs server | grep -i migrat
```

Expected: Migration logs showing 7 migrations applied

**Step 4: Test invite flow manually**

Run:
```bash
# Login as admin
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.local","password":"Test1234"}'

# Create an invite (use token from login)
curl -X POST http://localhost:8080/api/v1/teams/a0000000-0000-0000-0000-000000000001/invites \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"max_uses": 5}'

# List invites
curl http://localhost:8080/api/v1/teams/a0000000-0000-0000-0000-000000000001/invites \
  -H "Authorization: Bearer <token>"
```

**Step 5: Commit final state**

```bash
git add -A
git commit -m "feat: complete team invites implementation"
```

---

## Summary

19 tasks covering:
1. **Tasks 1-6**: Migration consolidation (23 → 7 migrations)
2. **Task 7**: New team_invites migration
3. **Task 8**: TeamInvite domain entity with tests
4. **Task 9**: TeamInviteDB adapter
5. **Tasks 10-11**: Repository and service extensions
6. **Tasks 12-14**: Handler implementations
7. **Task 15**: Teams service with invite logic
8. **Tasks 16-18**: Wiring and initialization
9. **Task 19**: Testing and verification

Each task is atomic with TDD where applicable, exact file paths, and commit points.
