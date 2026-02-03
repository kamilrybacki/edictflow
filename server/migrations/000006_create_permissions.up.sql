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
    ('a0000001-0000-0000-0000-00000000000b', 'view_audit_log', 'View change history', 'admin');
