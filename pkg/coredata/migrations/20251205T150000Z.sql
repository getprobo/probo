ALTER TABLE sessions ADD COLUMN tenant_id TEXT;
ALTER TABLE sessions ADD COLUMN parent_session_id TEXT REFERENCES sessions(id);

ALTER TABLE sessions ADD CONSTRAINT session_tenant_check CHECK (
    (parent_session_id IS NULL AND tenant_id IS NULL) OR
    (parent_session_id IS NOT NULL AND tenant_id IS NOT NULL)
);

