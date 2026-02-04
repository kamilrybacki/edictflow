-- 000022_extend_rules_three_file.down.sql

-- Remove indexes
DROP INDEX IF EXISTS idx_rules_category_id;
DROP INDEX IF EXISTS idx_rules_effective_dates;
DROP INDEX IF EXISTS idx_rules_target_teams;
DROP INDEX IF EXISTS idx_rules_target_users;
DROP INDEX IF EXISTS idx_rules_tags;

-- Remove constraint
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_target_layer_check;

-- Restore original target_layer values
UPDATE rules SET target_layer = 'global' WHERE target_layer = 'user';
UPDATE rules SET target_layer = 'local' WHERE target_layer = 'project';

-- Remove new columns
ALTER TABLE rules DROP COLUMN IF EXISTS description;
ALTER TABLE rules DROP COLUMN IF EXISTS category_id;
ALTER TABLE rules DROP COLUMN IF EXISTS overridable;
ALTER TABLE rules DROP COLUMN IF EXISTS effective_start;
ALTER TABLE rules DROP COLUMN IF EXISTS effective_end;
ALTER TABLE rules DROP COLUMN IF EXISTS target_teams;
ALTER TABLE rules DROP COLUMN IF EXISTS target_users;
ALTER TABLE rules DROP COLUMN IF EXISTS tags;
