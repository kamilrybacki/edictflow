CREATE TABLE rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    target_layer VARCHAR(50) NOT NULL,
    priority_weight INTEGER NOT NULL DEFAULT 0,
    triggers JSONB NOT NULL DEFAULT '[]',
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rules_team_id ON rules(team_id);
CREATE INDEX idx_rules_target_layer ON rules(target_layer);
