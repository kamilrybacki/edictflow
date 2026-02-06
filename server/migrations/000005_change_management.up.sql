-- 000005_change_management.up.sql
-- Change management: change_requests, exception_requests, notifications, notification_channels

-- Change requests
CREATE TABLE change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES rules(id),
    agent_id UUID NOT NULL REFERENCES agents(id),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID NOT NULL REFERENCES teams(id),
    file_path TEXT NOT NULL,
    original_hash TEXT NOT NULL,
    modified_hash TEXT NOT NULL,
    diff_content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    enforcement_mode TEXT NOT NULL,
    timeout_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

CREATE INDEX idx_change_requests_team_status ON change_requests(team_id, status);
CREATE INDEX idx_change_requests_rule_id ON change_requests(rule_id);
CREATE INDEX idx_change_requests_agent_id ON change_requests(agent_id);
CREATE INDEX idx_change_requests_timeout_at ON change_requests(timeout_at);

-- Exception requests
CREATE TABLE exception_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id UUID NOT NULL REFERENCES change_requests(id),
    user_id UUID NOT NULL REFERENCES users(id),
    justification TEXT NOT NULL,
    exception_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

CREATE INDEX idx_exception_requests_change_request_id ON exception_requests(change_request_id);
CREATE INDEX idx_exception_requests_status ON exception_requests(status);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID REFERENCES teams(id),
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_read_at ON notifications(user_id, read_at);
CREATE INDEX idx_notifications_team_id ON notifications(team_id);

-- Notification channels
CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    channel_type TEXT NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_channels_team_enabled ON notification_channels(team_id, enabled);
