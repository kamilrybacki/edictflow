-- Make team_id nullable for global rules
ALTER TABLE rules ALTER COLUMN team_id DROP NOT NULL;

-- Add force column for global rules
ALTER TABLE rules ADD COLUMN force BOOLEAN NOT NULL DEFAULT false;

-- Constraint: force only valid for global rules (team_id IS NULL)
ALTER TABLE rules ADD CONSTRAINT rules_force_global_only
    CHECK (force = false OR team_id IS NULL);

-- Constraint: global rules must be enterprise layer
ALTER TABLE rules ADD CONSTRAINT rules_global_enterprise_only
    CHECK (team_id IS NOT NULL OR target_layer = 'enterprise');

-- Index for efficient global rule queries
CREATE INDEX idx_rules_global ON rules(force) WHERE team_id IS NULL;

-- Index for force flag queries
CREATE INDEX idx_rules_force ON rules(force) WHERE force = true;
