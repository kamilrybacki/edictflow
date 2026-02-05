-- Remove indexes
DROP INDEX IF EXISTS idx_rules_force;
DROP INDEX IF EXISTS idx_rules_global;

-- Remove constraints
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_global_enterprise_only;
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_force_global_only;

-- Remove force column
ALTER TABLE rules DROP COLUMN IF EXISTS force;

-- Note: Cannot restore NOT NULL on team_id if global rules exist
-- Must delete global rules first: DELETE FROM rules WHERE team_id IS NULL;
ALTER TABLE rules ALTER COLUMN team_id SET NOT NULL;
