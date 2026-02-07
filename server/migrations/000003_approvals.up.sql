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
    ('c0000001-0000-0000-0000-000000000001', 'project', 'approve_project', 1),
    ('c0000001-0000-0000-0000-000000000002', 'team', 'approve_team', 1),
    ('c0000001-0000-0000-0000-000000000003', 'organization', 'approve_organization', 2);

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
