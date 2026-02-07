-- Restore team ownership from attachments

-- Step 1: Restore team_id from first attachment for each rule
UPDATE rules r
SET team_id = (
    SELECT ra.team_id
    FROM rule_attachments ra
    WHERE ra.rule_id = r.id
    ORDER BY ra.created_at ASC
    LIMIT 1
);

-- Step 2: Delete all attachments
DELETE FROM rule_attachments;

-- Step 3: Restore the constraint
ALTER TABLE rules ADD CONSTRAINT rules_global_organization_only
    CHECK (team_id IS NOT NULL OR target_layer = 'organization');
