ALTER TABLE rules DROP COLUMN status;
ALTER TABLE rules DROP COLUMN created_by;
ALTER TABLE rules DROP COLUMN submitted_at;
ALTER TABLE rules DROP COLUMN approved_at;
DROP INDEX IF EXISTS idx_rules_status;
DROP INDEX IF EXISTS idx_rules_created_by;
