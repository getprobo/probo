CREATE TABLE log_exports (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    from_time TIMESTAMPTZ NOT NULL,
    to_time TIMESTAMPTZ NOT NULL,
    file_id TEXT,
    error TEXT,
    recipient_email TEXT NOT NULL,
    recipient_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_log_exports_pending ON log_exports (status, created_at)
    WHERE status = 'PENDING';
