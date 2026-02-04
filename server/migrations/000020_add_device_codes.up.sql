-- 000020_add_device_codes.up.sql
CREATE TABLE device_codes (
    device_code TEXT PRIMARY KEY,
    user_code TEXT UNIQUE NOT NULL,
    user_id UUID REFERENCES users(id),
    client_id TEXT NOT NULL DEFAULT 'claudeception-cli',
    expires_at TIMESTAMPTZ NOT NULL,
    authorized_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_device_codes_user_code ON device_codes(user_code);
CREATE INDEX idx_device_codes_expires_at ON device_codes(expires_at);
