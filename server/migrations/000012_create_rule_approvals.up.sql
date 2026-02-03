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
