-- 000006_extensions.down.sql
ALTER TABLE rules DROP CONSTRAINT IF EXISTS fk_rules_category;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS device_codes;
