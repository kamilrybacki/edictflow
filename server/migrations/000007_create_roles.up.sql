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
