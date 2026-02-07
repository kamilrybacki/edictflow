-- Migrate existing team rules to library model with attachments

-- Step 0: Drop the constraint that prevents library rules with non-organization layers
-- In the library model, rules can have any target layer regardless of team_id
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_global_organization_only;

-- Step 1: Create attachments for existing team rules
INSERT INTO rule_attachments (id, rule_id, team_id, enforcement_mode, temporary_timeout_hours, status, requested_by, approved_by, created_at, approved_at)
SELECT
    gen_random_uuid(),
    r.id,
    r.team_id,
    r.enforcement_mode,
    r.temporary_timeout_hours,
    CASE WHEN r.status = 'approved' THEN 'approved' ELSE 'pending' END,
    COALESCE(r.created_by, (SELECT id FROM users LIMIT 1)),
    CASE WHEN r.status = 'approved' THEN r.created_by ELSE NULL END,
    r.created_at,
    CASE WHEN r.status = 'approved' THEN r.approved_at ELSE NULL END
FROM rules r
WHERE r.team_id IS NOT NULL;

-- Step 2: Create attachments for enterprise/forced rules (attach to all teams)
INSERT INTO rule_attachments (id, rule_id, team_id, enforcement_mode, temporary_timeout_hours, status, requested_by, approved_by, created_at, approved_at)
SELECT
    gen_random_uuid(),
    r.id,
    t.id,
    r.enforcement_mode,
    r.temporary_timeout_hours,
    'approved',
    COALESCE(r.created_by, (SELECT id FROM users LIMIT 1)),
    r.created_by,
    r.created_at,
    r.approved_at
FROM rules r
CROSS JOIN teams t
WHERE r.team_id IS NULL AND r.status = 'approved';

-- Step 3: Nullify team_id on all rules (they're now library rules)
-- Note: We don't drop the column yet to allow rollback
UPDATE rules SET team_id = NULL;
