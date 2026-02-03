CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- Assign permissions to Member role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('b0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001'), -- create_rules
    ('b0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000004'); -- approve_local

-- Assign all permissions to Admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000001-0000-0000-0000-000000000002', id FROM permissions;
