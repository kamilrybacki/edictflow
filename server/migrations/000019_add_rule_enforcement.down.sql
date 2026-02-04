ALTER TABLE rules
DROP COLUMN IF EXISTS enforcement_mode,
DROP COLUMN IF EXISTS temporary_timeout_hours;
