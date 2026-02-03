# User Management, RBAC & Rule Approval System Design

**Date:** 2026-02-03
**Status:** Approved for implementation

## Overview

This design adds user management with local authentication, a flexible role-based access control (RBAC) system with inheritance, a quorum-based rule approval workflow, and a full audit log. The admin panel allows configuration of roles, permissions, and approval requirements.

## Key Decisions

| Decision | Choice |
|----------|--------|
| Authentication | Local auth (email/password with bcrypt) |
| Role hierarchy | Configurable via admin panel |
| Rule pending state | Draft (invisible to agents until approved) |
| Approval mechanism | Quorum (N approvals from users with required permission) |
| RBAC complexity | Full RBAC with role inheritance |
| Audit logging | Full history with change diffs |
| Approval config | Hierarchical (global defaults, team can only tighten) |
| Session management | Simple JWT (24h expiry) |
| Password policy | 8+ chars, uppercase, lowercase, number |

---

## 1. User & Authentication System

### User Model Changes

Add to existing User domain:

- `password_hash` - bcrypt hash (cost 12)
- `email_verified` - boolean (default true for now)
- `created_by` - UUID of user who created this user (null for self-registered)
- `last_login_at` - timestamp
- `is_active` - boolean for soft delete

### Endpoints

```
POST /api/v1/auth/register     - Create account (email, name, password)
POST /api/v1/auth/login        - Returns JWT (24h expiry)
POST /api/v1/auth/logout       - Invalidate token
GET  /api/v1/auth/me           - Get current user profile
PUT  /api/v1/auth/me           - Update own profile (name, password)
```

### Password Validation

- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number

### JWT Claims

```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "team_id": "team-uuid",
  "permissions": ["create_rules", "approve_global"],
  "exp": 1234567890
}
```

---

## 2. Role-Based Access Control (RBAC)

### Models

**Permission:**
```
Permission {
  id: uuid
  code: string              // "create_rules", "approve_global"
  description: string
  category: string          // "rules", "teams", "users", "admin"
}
```

**Role:**
```
Role {
  id: uuid
  name: string              // "Lead", "Manager", "Admin"
  description: string
  hierarchy_level: int      // 1=lowest, higher=more senior
  parent_role_id: uuid?     // Inherits permissions from parent
  team_id: uuid?            // null = global role, set = team-specific
  is_system: bool           // true = cannot be deleted
  created_at: timestamp
}
```

**Role-Permission Mapping:**
```
RolePermission {
  role_id: uuid
  permission_id: uuid
}
```

**User-Role Assignment:**
```
UserRole {
  user_id: uuid
  role_id: uuid
  assigned_by: uuid
  assigned_at: timestamp
}
```

### Built-in Permissions

| Code | Category | Description |
|------|----------|-------------|
| `create_rules` | rules | Create new rules (as drafts) |
| `edit_rules` | rules | Edit existing rules |
| `delete_rules` | rules | Delete rules |
| `approve_local` | rules | Approve local-scoped rules |
| `approve_project` | rules | Approve project-scoped rules |
| `approve_global` | rules | Approve global-scoped rules |
| `approve_enterprise` | rules | Approve enterprise-scoped rules |
| `manage_users` | users | Create/edit/deactivate users |
| `manage_roles` | admin | Create/edit roles and assign permissions |
| `manage_team_settings` | teams | Edit team configuration |
| `view_audit_log` | admin | View change history |

### Default System Roles

- **Member** (level 1): `create_rules`, `approve_local`
- **Admin** (level 100): All permissions, cannot be deleted

---

## 3. Rule Approval Workflow

### Rule Status

```
draft      → Newly created, not visible to agents
pending    → Submitted for approval, awaiting quorum
approved   → Fully approved, active and visible to agents
rejected   → Approval denied, author can revise and resubmit
```

### Rule Model Changes

Add to existing Rule domain:

- `status` - RuleStatus (draft, pending, approved, rejected)
- `created_by` - UUID of author
- `submitted_at` - timestamp when submitted for approval
- `approved_at` - timestamp when fully approved

### Approval Model

```
RuleApproval {
  id: uuid
  rule_id: uuid
  user_id: uuid               // Who approved/rejected
  decision: string            // "approved" or "rejected"
  comment: string?            // Optional feedback
  created_at: timestamp
}
```

### Approval Config Model

```
ApprovalConfig {
  id: uuid
  scope: TargetLayer          // local, project, global, enterprise
  required_permission: string // e.g., "approve_global"
  required_count: int         // Quorum size
  team_id: uuid?              // null = global default, set = team override
}
```

### Approval Flow

1. User creates rule → status = `draft`
2. User submits rule → status = `pending`, `submitted_at` = now
3. Users with required permission can approve or reject
4. When approval count >= required_count → status = `approved`
5. If any rejection → status = `rejected`, author notified
6. Author can edit rejected rule and resubmit

### Hierarchical Config Logic

- System loads global default for scope
- If team has override for that scope, use team's `required_count` (must be >= global)
- Team cannot lower requirements below global defaults

### Endpoints

```
POST /api/v1/rules/{id}/submit      - Submit draft for approval
POST /api/v1/rules/{id}/approve     - Add approval (with optional comment)
POST /api/v1/rules/{id}/reject      - Reject (requires comment)
GET  /api/v1/rules/{id}/approvals   - List approvals for a rule
```

---

## 4. Audit Log

### Audit Entry Model

```
AuditEntry {
  id: uuid
  entity_type: string         // "rule", "user", "role", "team", "approval_config"
  entity_id: uuid             // ID of affected entity
  action: string              // "created", "updated", "deleted", "submitted", "approved", "rejected"
  actor_id: uuid              // User who performed action
  changes: jsonb              // {"field": {"old": x, "new": y}, ...}
  metadata: jsonb?            // Additional context (IP, user agent, etc.)
  created_at: timestamp
}
```

### What Gets Logged

| Entity | Actions |
|--------|---------|
| Rule | created, updated, deleted, submitted, approved, rejected |
| User | created, updated, deactivated, role_assigned, role_removed |
| Role | created, updated, deleted, permission_added, permission_removed |
| Team | created, updated, deleted |
| ApprovalConfig | created, updated, deleted |

### Changes Field Example

```json
{
  "name": {"old": "Old Rule Name", "new": "New Rule Name"},
  "content": {"old": "...", "new": "..."}
}
```

### Endpoints

```
GET /api/v1/audit                     - List audit entries (paginated, filterable)
GET /api/v1/audit/entity/{type}/{id}  - History for specific entity
```

### Query Filters

- `entity_type` - Filter by type
- `entity_id` - Filter by specific entity
- `actor_id` - Filter by who made changes
- `action` - Filter by action type
- `from` / `to` - Date range

### Retention

- Audit logs are immutable (no updates/deletes)
- Optional: admin-configurable retention period (default: keep forever)

---

## 5. Admin Panel

### Sections

**1. User Management** (`/admin/users`)
- List all users with search/filter by team, role, status
- Create new user (assigns to team, sets initial role)
- Edit user (name, email, team assignment)
- Deactivate/reactivate user (soft delete)
- Assign/remove roles from user

**2. Role Management** (`/admin/roles`)
- List all roles (system + custom)
- Create new role: name, description, hierarchy level, parent role, scope
- Edit role permissions (checkbox grid: role × permission)
- Delete custom roles (must reassign users first)
- View inherited permissions (from parent role)

**3. Approval Configuration** (`/admin/approvals`)
- Grid view: Scope × Team showing required count
- Edit global defaults (top row)
- Per-team overrides (can only increase, not decrease)
- Visual indicator when team differs from global

**4. Audit Log Viewer** (`/admin/audit`)
- Paginated table of audit entries
- Filters: entity type, action, actor, date range
- Click entry to see full change diff

### Access Control

- Entire admin panel requires `manage_roles` or `manage_users` permission
- Role management requires `manage_roles`
- User management requires `manage_users`
- Audit log requires `view_audit_log`

---

## 6. Dashboard UI Integration

### Header Changes

- Add user menu (top right): avatar, name, dropdown with "Profile", "Logout"
- Show current user's primary role as badge

### Login/Register Pages

- `/login` - Email, password, "Register" link
- `/register` - Email, name, password, confirm password
- Redirect to `/` (dashboard) after successful auth
- Protected routes redirect to `/login` if no valid token

### TeamList Sidebar Enhancements

- Show user's role within selected team
- Admin users see "Admin Panel" link at bottom

### RuleList Changes

- Each rule shows:
  - Status badge: Draft (gray), Pending (yellow), Approved (green), Rejected (red)
  - Author name and avatar
  - "Pending X of Y approvals" indicator for pending rules
- Filter tabs: All | My Rules | Pending Approval | Approved
- "Submit for Approval" button on draft rules

### Rule Detail/Editor Modal

- Show author info (created by, created at)
- Show approval history (who approved/rejected, when, comments)
- For pending rules user can approve: "Approve" and "Reject" buttons
- Comment field (required for rejection)
- For rejected rules: show rejection reason, "Edit & Resubmit" button

### New Admin Routes

- `/admin/users` - User management
- `/admin/roles` - Role & permission management
- `/admin/approvals` - Approval config
- `/admin/audit` - Audit log viewer

### Permission-Based UI

- Hide "Create Rule" button if user lacks `create_rules`
- Hide "Delete" buttons if user lacks `delete_rules`
- Hide approval buttons if user lacks required approval permission
- Disable admin sections user cannot access

---

## 7. Database Schema

### New Tables

```sql
-- Users with password auth
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(500),
  email_verified BOOLEAN DEFAULT true,
  team_id UUID REFERENCES teams(id),
  created_by UUID REFERENCES users(id),
  last_login_at TIMESTAMP WITH TIME ZONE,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Permissions (seeded, rarely changed)
CREATE TABLE permissions (
  id UUID PRIMARY KEY,
  code VARCHAR(100) UNIQUE NOT NULL,
  description TEXT,
  category VARCHAR(50) NOT NULL
);

-- Roles with hierarchy and inheritance
CREATE TABLE roles (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  hierarchy_level INT NOT NULL,
  parent_role_id UUID REFERENCES roles(id),
  team_id UUID REFERENCES teams(id),
  is_system BOOLEAN DEFAULT false,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(name, team_id)
);

-- Role-permission assignments
CREATE TABLE role_permissions (
  role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

-- User-role assignments
CREATE TABLE user_roles (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
  assigned_by UUID REFERENCES users(id),
  assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  PRIMARY KEY (user_id, role_id)
);

-- Approval configuration per scope
CREATE TABLE approval_configs (
  id UUID PRIMARY KEY,
  scope VARCHAR(50) NOT NULL,
  required_permission VARCHAR(100) NOT NULL,
  required_count INT NOT NULL DEFAULT 1,
  team_id UUID REFERENCES teams(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(scope, team_id)
);

-- Rule approvals
CREATE TABLE rule_approvals (
  id UUID PRIMARY KEY,
  rule_id UUID REFERENCES rules(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id),
  decision VARCHAR(20) NOT NULL,
  comment TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit log (append-only)
CREATE TABLE audit_entries (
  id UUID PRIMARY KEY,
  entity_type VARCHAR(50) NOT NULL,
  entity_id UUID NOT NULL,
  action VARCHAR(50) NOT NULL,
  actor_id UUID REFERENCES users(id),
  changes JSONB,
  metadata JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Modifications to Existing Tables

```sql
ALTER TABLE rules ADD COLUMN status VARCHAR(20) DEFAULT 'draft';
ALTER TABLE rules ADD COLUMN created_by UUID REFERENCES users(id);
ALTER TABLE rules ADD COLUMN submitted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE rules ADD COLUMN approved_at TIMESTAMP WITH TIME ZONE;
```

### Indexes

```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_team ON users(team_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_rule_approvals_rule ON rule_approvals(rule_id);
CREATE INDEX idx_audit_entity ON audit_entries(entity_type, entity_id);
CREATE INDEX idx_audit_actor ON audit_entries(actor_id);
CREATE INDEX idx_audit_created ON audit_entries(created_at);
```

---

## 8. API Endpoints Summary

### Authentication

```
POST /api/v1/auth/register     - Create account
POST /api/v1/auth/login        - Login, returns JWT
POST /api/v1/auth/logout       - Invalidate token
GET  /api/v1/auth/me           - Current user profile
PUT  /api/v1/auth/me           - Update own profile/password
```

### Users (admin)

```
GET    /api/v1/users           - List users (filterable)
POST   /api/v1/users           - Create user
GET    /api/v1/users/{id}      - Get user
PUT    /api/v1/users/{id}      - Update user
DELETE /api/v1/users/{id}      - Deactivate user
POST   /api/v1/users/{id}/roles/{roleId}    - Assign role
DELETE /api/v1/users/{id}/roles/{roleId}    - Remove role
```

### Roles (admin)

```
GET    /api/v1/roles           - List roles
POST   /api/v1/roles           - Create role
GET    /api/v1/roles/{id}      - Get role with permissions
PUT    /api/v1/roles/{id}      - Update role
DELETE /api/v1/roles/{id}      - Delete role
GET    /api/v1/permissions     - List all permissions
```

### Approval Config (admin)

```
GET    /api/v1/approval-configs              - List all configs
PUT    /api/v1/approval-configs/{scope}      - Update global default
PUT    /api/v1/approval-configs/{scope}/teams/{teamId}  - Set team override
DELETE /api/v1/approval-configs/{scope}/teams/{teamId}  - Remove team override
```

### Rules (modified)

```
POST   /api/v1/rules                  - Create rule (status=draft)
PUT    /api/v1/rules/{id}             - Update rule
POST   /api/v1/rules/{id}/submit      - Submit for approval
POST   /api/v1/rules/{id}/approve     - Approve rule
POST   /api/v1/rules/{id}/reject      - Reject rule (comment required)
GET    /api/v1/rules/{id}/approvals   - Get approval history
```

### Audit (admin)

```
GET    /api/v1/audit                        - List entries (paginated)
GET    /api/v1/audit/entity/{type}/{id}     - Entity history
```

---

## Implementation Order

1. **Database migrations** - Create all new tables, modify rules table
2. **Domain models** - Add Go structs for new entities
3. **Auth system** - Register, login, JWT generation, password validation
4. **User CRUD** - User repository, service, handlers
5. **RBAC core** - Permissions, roles, role-permissions, user-roles
6. **Permission middleware** - Check permissions on protected endpoints
7. **Rule approval** - Status field, approval workflow, approval config
8. **Audit logging** - Service that logs all changes
9. **Frontend auth** - Login/register pages, token storage, protected routes
10. **Admin panel** - User, role, approval config, audit UI
11. **Dashboard updates** - Rule status, author display, approval UI
