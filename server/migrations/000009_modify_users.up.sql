-- Add new columns for local auth
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT true;

-- Make team_id optional (null for system-wide admins)
ALTER TABLE users ALTER COLUMN team_id DROP NOT NULL;

-- Drop the old role column (will use user_roles junction)
ALTER TABLE users DROP COLUMN role;

CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_created_by ON users(created_by);
