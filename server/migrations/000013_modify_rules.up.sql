ALTER TABLE rules ADD COLUMN status VARCHAR(20) DEFAULT 'draft'
    CHECK (status IN ('draft', 'pending', 'approved', 'rejected'));
ALTER TABLE rules ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE rules ADD COLUMN submitted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE rules ADD COLUMN approved_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_rules_status ON rules(status);
CREATE INDEX idx_rules_created_by ON rules(created_by);
