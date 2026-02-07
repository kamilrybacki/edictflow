-- Seed data for test infrastructure
-- This creates test users, team, and rules for manual testing
--
-- Test Accounts:
--   user@test.local   / Test1234  (Member role)
--   admin@test.local  / Test1234  (Admin role)
--   alex.rivera@test.local   / Test1234  (Member role) - Auto-connected agent
--   jordan.kim@test.local    / Test1234  (Member role) - Auto-connected agent
--   sarah.chen@test.local    / Test1234  (Member role) - Auto-connected agent
--   mike.johnson@test.local  / Test1234  (Member role) - Auto-connected agent
--   emma.wilson@test.local   / Test1234  (Member role) - Auto-connected agent

-- Create test team
INSERT INTO teams (id, name, settings, created_at)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Test Team',
    '{}'::jsonb,
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Password hash for 'Test1234' generated with bcrypt cost 12
-- To regenerate: go run -mod=mod server/... with bcrypt.GenerateFromPassword

-- Create regular user (Member)
INSERT INTO users (id, email, name, auth_provider, team_id, is_active, password_hash, email_verified, created_at)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'user@test.local',
    'Test User',
    'local',
    'a0000000-0000-0000-0000-000000000001',
    true,
    '$2a$12$/29KgikA2vLGIyn4MVMbO.F5bgBFu0JXXcdmyX7MZDd7T1BvbKxhm',
    true,
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash;

-- Create admin user
INSERT INTO users (id, email, name, auth_provider, team_id, is_active, password_hash, email_verified, created_at)
VALUES (
    'c0000000-0000-0000-0000-000000000002',
    'admin@test.local',
    'Test Admin',
    'local',
    'a0000000-0000-0000-0000-000000000001',
    true,
    '$2a$12$/29KgikA2vLGIyn4MVMbO.F5bgBFu0JXXcdmyX7MZDd7T1BvbKxhm',
    true,
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash;

-- Assign Member role to regular user
-- Role ID b0000001-0000-0000-0000-000000000001 = Member
INSERT INTO user_roles (user_id, role_id, assigned_at)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'b0000001-0000-0000-0000-000000000001',
    NOW()
) ON CONFLICT (user_id, role_id) DO NOTHING;

-- Assign Admin role to admin user
-- Role ID b0000001-0000-0000-0000-000000000002 = Admin
INSERT INTO user_roles (user_id, role_id, assigned_at)
VALUES (
    'c0000000-0000-0000-0000-000000000002',
    'b0000001-0000-0000-0000-000000000002',
    NOW()
) ON CONFLICT (user_id, role_id) DO NOTHING;

-- ============================================
-- Agent Users (for auto-connected test clients)
-- ============================================

-- Alex Rivera
INSERT INTO users (id, email, name, auth_provider, team_id, is_active, password_hash, email_verified, created_at)
VALUES (
    'c0000000-0000-0000-0000-000000000011',
    'alex.rivera@test.local',
    'Alex Rivera',
    'local',
    'a0000000-0000-0000-0000-000000000001',
    true,
    '$2a$12$/29KgikA2vLGIyn4MVMbO.F5bgBFu0JXXcdmyX7MZDd7T1BvbKxhm',
    true,
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash;

INSERT INTO user_roles (user_id, role_id, assigned_at)
VALUES (
    'c0000000-0000-0000-0000-000000000011',
    'b0000001-0000-0000-0000-000000000001',
    NOW()
) ON CONFLICT (user_id, role_id) DO NOTHING;

-- Jordan Kim
INSERT INTO users (id, email, name, auth_provider, team_id, is_active, password_hash, email_verified, created_at)
VALUES (
    'c0000000-0000-0000-0000-000000000012',
    'jordan.kim@test.local',
    'Jordan Kim',
    'local',
    'a0000000-0000-0000-0000-000000000001',
    true,
    '$2a$12$/29KgikA2vLGIyn4MVMbO.F5bgBFu0JXXcdmyX7MZDd7T1BvbKxhm',
    true,
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash;

INSERT INTO user_roles (user_id, role_id, assigned_at)
VALUES (
    'c0000000-0000-0000-0000-000000000012',
    'b0000001-0000-0000-0000-000000000001',
    NOW()
) ON CONFLICT (user_id, role_id) DO NOTHING;

-- Sarah Chen
INSERT INTO users (id, email, name, auth_provider, team_id, is_active, password_hash, email_verified, created_at)
VALUES (
    'c0000000-0000-0000-0000-000000000013',
    'sarah.chen@test.local',
    'Sarah Chen',
    'local',
    'a0000000-0000-0000-0000-000000000001',
    true,
    '$2a$12$/29KgikA2vLGIyn4MVMbO.F5bgBFu0JXXcdmyX7MZDd7T1BvbKxhm',
    true,
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash;

INSERT INTO user_roles (user_id, role_id, assigned_at)
VALUES (
    'c0000000-0000-0000-0000-000000000013',
    'b0000001-0000-0000-0000-000000000001',
    NOW()
) ON CONFLICT (user_id, role_id) DO NOTHING;

-- Mike Johnson
INSERT INTO users (id, email, name, auth_provider, team_id, is_active, password_hash, email_verified, created_at)
VALUES (
    'c0000000-0000-0000-0000-000000000014',
    'mike.johnson@test.local',
    'Mike Johnson',
    'local',
    'a0000000-0000-0000-0000-000000000001',
    true,
    '$2a$12$/29KgikA2vLGIyn4MVMbO.F5bgBFu0JXXcdmyX7MZDd7T1BvbKxhm',
    true,
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash;

INSERT INTO user_roles (user_id, role_id, assigned_at)
VALUES (
    'c0000000-0000-0000-0000-000000000014',
    'b0000001-0000-0000-0000-000000000001',
    NOW()
) ON CONFLICT (user_id, role_id) DO NOTHING;

-- Emma Wilson
INSERT INTO users (id, email, name, auth_provider, team_id, is_active, password_hash, email_verified, created_at)
VALUES (
    'c0000000-0000-0000-0000-000000000015',
    'emma.wilson@test.local',
    'Emma Wilson',
    'local',
    'a0000000-0000-0000-0000-000000000001',
    true,
    '$2a$12$/29KgikA2vLGIyn4MVMbO.F5bgBFu0JXXcdmyX7MZDd7T1BvbKxhm',
    true,
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash;

INSERT INTO user_roles (user_id, role_id, assigned_at)
VALUES (
    'c0000000-0000-0000-0000-000000000015',
    'b0000001-0000-0000-0000-000000000001',
    NOW()
) ON CONFLICT (user_id, role_id) DO NOTHING;

-- Create test rule with block mode
INSERT INTO rules (id, name, team_id, content, target_layer, priority_weight, triggers, enforcement_mode, created_at, updated_at)
VALUES (
    'd0000000-0000-0000-0000-000000000001',
    'Standard CLAUDE.md',
    'a0000000-0000-0000-0000-000000000001',
    '# CLAUDE.md

This is the **approved** content managed by Edictflow.

## Guidelines

- Follow best practices
- Write tests for all features
- Keep functions under 50 lines

## Do Not Modify

This file is managed centrally. Changes will be reverted.',
    'project',
    100,
    '[{"type": "path", "pattern": "CLAUDE.md"}, {"type": "glob", "pattern": "**/CLAUDE.md"}]'::jsonb,
    'block',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Create a second rule with warning mode for testing
INSERT INTO rules (id, name, team_id, content, target_layer, priority_weight, triggers, enforcement_mode, created_at, updated_at)
VALUES (
    'd0000000-0000-0000-0000-000000000002',
    'Guidelines (Warning)',
    'a0000000-0000-0000-0000-000000000001',
    '# Guidelines

These are optional guidelines.',
    'project',
    50,
    '[{"type": "path", "pattern": "GUIDELINES.md"}]'::jsonb,
    'warning',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- ============================================
-- Test Team Invite
-- ============================================

INSERT INTO team_invites (id, team_id, code, max_uses, use_count, expires_at, created_by, created_at)
VALUES (
    'e0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'TESTCODE',
    100,
    0,
    NOW() + INTERVAL '30 days',
    'c0000000-0000-0000-0000-000000000002',
    NOW()
) ON CONFLICT (id) DO NOTHING;
