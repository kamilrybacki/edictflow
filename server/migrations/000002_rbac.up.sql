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
