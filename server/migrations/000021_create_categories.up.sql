-- 000021_create_categories.up.sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    org_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Unique constraint: name must be unique within org (or globally for system categories)
CREATE UNIQUE INDEX idx_categories_name_org ON categories(name, COALESCE(org_id, '00000000-0000-0000-0000-000000000000'));

-- Insert system default categories
INSERT INTO categories (id, name, is_system, org_id, display_order) VALUES
    (gen_random_uuid(), 'Security', TRUE, NULL, 1),
    (gen_random_uuid(), 'Coding Standards', TRUE, NULL, 2),
    (gen_random_uuid(), 'Testing', TRUE, NULL, 3),
    (gen_random_uuid(), 'Documentation', TRUE, NULL, 4);
