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
    status TEXT NOT NULL DEFAULT 'pending',
    enforcement_mode TEXT NOT NULL,
    timeout_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

-- Indexes for efficient queries
CREATE INDEX idx_change_requests_team_status ON change_requests(team_id, status);
CREATE INDEX idx_change_requests_rule_id ON change_requests(rule_id);
CREATE INDEX idx_change_requests_agent_id ON change_requests(agent_id);
CREATE INDEX idx_change_requests_timeout_at ON change_requests(timeout_at);
