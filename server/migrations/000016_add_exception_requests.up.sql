-- Exception requests for rejected changes
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

-- Indexes for efficient queries
CREATE INDEX idx_exception_requests_change_request_id ON exception_requests(change_request_id);
CREATE INDEX idx_exception_requests_status ON exception_requests(status);
