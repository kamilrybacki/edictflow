-- 000006_extensions.up.sql
-- Extensions: device_codes, categories (rules extensions already in 000001)

-- Device codes for CLI auth
CREATE TABLE device_codes (
    device_code TEXT PRIMARY KEY,
    user_code TEXT UNIQUE NOT NULL,
    user_id UUID REFERENCES users(id),
    client_id TEXT NOT NULL DEFAULT 'edictflow-cli',
    expires_at TIMESTAMPTZ NOT NULL,
    authorized_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_device_codes_user_code ON device_codes(user_code);
CREATE INDEX idx_device_codes_expires_at ON device_codes(expires_at);

-- Categories for rules
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    org_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_categories_name_org ON categories(name, COALESCE(org_id, '00000000-0000-0000-0000-000000000000'));

-- Insert system default categories
INSERT INTO categories (id, name, is_system, org_id, display_order) VALUES
    (gen_random_uuid(), 'Security', TRUE, NULL, 1),
    (gen_random_uuid(), 'Coding Standards', TRUE, NULL, 2),
    (gen_random_uuid(), 'Testing', TRUE, NULL, 3),
    (gen_random_uuid(), 'Documentation', TRUE, NULL, 4);

-- Add foreign key from rules to categories (rules table already has category_id column)
ALTER TABLE rules ADD CONSTRAINT fk_rules_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
