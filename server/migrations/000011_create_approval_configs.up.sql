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
