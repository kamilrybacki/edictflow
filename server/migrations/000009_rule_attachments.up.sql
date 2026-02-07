-- Rule attachments: link library rules to teams with enforcement settings

CREATE TABLE rule_attachments (
    id UUID PRIMARY KEY,
    rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    enforcement_mode TEXT NOT NULL DEFAULT 'block' CHECK (enforcement_mode IN ('block', 'temporary', 'warning')),
    temporary_timeout_hours INTEGER NOT NULL DEFAULT 24,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(rule_id, team_id)
);

CREATE INDEX idx_rule_attachments_rule_id ON rule_attachments(rule_id);
CREATE INDEX idx_rule_attachments_team_id ON rule_attachments(team_id);
CREATE INDEX idx_rule_attachments_status ON rule_attachments(status);

-- Add approved_by column to rules table
ALTER TABLE rules ADD COLUMN approved_by UUID REFERENCES users(id) ON DELETE SET NULL;
