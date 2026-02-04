-- 000022_extend_rules_three_file.up.sql

-- Add new columns to rules table
ALTER TABLE rules ADD COLUMN description TEXT;
ALTER TABLE rules ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE SET NULL;
ALTER TABLE rules ADD COLUMN overridable BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE rules ADD COLUMN effective_start TIMESTAMP WITH TIME ZONE;
ALTER TABLE rules ADD COLUMN effective_end TIMESTAMP WITH TIME ZONE;
ALTER TABLE rules ADD COLUMN target_teams UUID[] DEFAULT '{}';
ALTER TABLE rules ADD COLUMN target_users UUID[] DEFAULT '{}';
ALTER TABLE rules ADD COLUMN tags TEXT[] DEFAULT '{}';

-- Update target_layer enum values: rename 'global' to 'user', remove 'local'
-- First update existing values
UPDATE rules SET target_layer = 'user' WHERE target_layer = 'global';
UPDATE rules SET target_layer = 'project' WHERE target_layer = 'local';

-- Add check constraint for new enum values
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_target_layer_check;
ALTER TABLE rules ADD CONSTRAINT rules_target_layer_check
    CHECK (target_layer IN ('enterprise', 'user', 'project'));

-- Create indexes for new columns
CREATE INDEX idx_rules_category_id ON rules(category_id);
CREATE INDEX idx_rules_effective_dates ON rules(effective_start, effective_end);
CREATE INDEX idx_rules_target_teams ON rules USING GIN(target_teams);
CREATE INDEX idx_rules_target_users ON rules USING GIN(target_users);
CREATE INDEX idx_rules_tags ON rules USING GIN(tags);

-- Assign existing rules to a default category (Coding Standards)
UPDATE rules SET category_id = (SELECT id FROM categories WHERE name = 'Coding Standards' AND is_system = TRUE LIMIT 1)
WHERE category_id IS NULL;
