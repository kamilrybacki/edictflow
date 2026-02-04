-- Add enforcement mode fields to existing rules table
ALTER TABLE rules
ADD COLUMN enforcement_mode TEXT NOT NULL DEFAULT 'block',
ADD COLUMN temporary_timeout_hours INTEGER NOT NULL DEFAULT 24;
