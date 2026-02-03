ALTER TABLE users ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'member';
ALTER TABLE users DROP COLUMN password_hash;
ALTER TABLE users DROP COLUMN email_verified;
ALTER TABLE users DROP COLUMN created_by;
ALTER TABLE users DROP COLUMN last_login_at;
ALTER TABLE users DROP COLUMN is_active;
ALTER TABLE users ALTER COLUMN team_id SET NOT NULL;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_created_by;
