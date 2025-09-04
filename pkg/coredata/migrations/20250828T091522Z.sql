CREATE TYPE framework_export_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

CREATE TABLE framework_exports (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    framework_id TEXT NOT NULL,
    status framework_export_status NOT NULL,
    file_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE framework_exports ADD CONSTRAINT framework_exports_framework_id_fkey
    FOREIGN KEY (framework_id)
    REFERENCES frameworks(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE framework_exports ADD CONSTRAINT framework_exports_file_id_fkey
    FOREIGN KEY (file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE SET NULL;
