// agent/storage/migrations.go
package storage

const schema = `
CREATE TABLE IF NOT EXISTS auth (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    user_email TEXT NOT NULL,
    user_name TEXT NOT NULL,
    team_id TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS message_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ref_id TEXT UNIQUE NOT NULL,
    msg_type TEXT NOT NULL,
    payload TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    attempts INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cached_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    content TEXT NOT NULL,
    description TEXT DEFAULT '',
    target_layer TEXT NOT NULL,
    category_id TEXT,
    category_name TEXT,
    overridable INTEGER DEFAULT 1,
    effective_start INTEGER,
    effective_end INTEGER,
    tags TEXT DEFAULT '[]',
    triggers TEXT NOT NULL,
    enforcement_mode TEXT NOT NULL,
    temporary_timeout_hours INTEGER,
    version INTEGER NOT NULL,
    cached_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS cached_categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_system INTEGER NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS watched_projects (
    path TEXT PRIMARY KEY,
    detected_context TEXT,
    detected_tags TEXT,
    last_sync_at INTEGER
);

CREATE TABLE IF NOT EXISTS pending_changes (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    original_content TEXT NOT NULL,
    modified_content TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Performance indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_cached_rules_layer ON cached_rules(target_layer);
CREATE INDEX IF NOT EXISTS idx_message_queue_attempts ON message_queue(attempts);
CREATE INDEX IF NOT EXISTS idx_pending_changes_status ON pending_changes(status);
CREATE INDEX IF NOT EXISTS idx_pending_changes_rule_id ON pending_changes(rule_id);
CREATE INDEX IF NOT EXISTS idx_watched_projects_last_sync ON watched_projects(last_sync_at);
`

func (s *Storage) migrate() error {
	_, err := s.db.Exec(schema)
	return err
}
