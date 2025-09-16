CREATE TYPE export_jobs_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);

CREATE TYPE export_jobs_type AS ENUM (
    'FRAMEWORK',
    'DOCUMENT'
);

CREATE TABLE export_jobs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    type export_jobs_type NOT NULL,
    arguments JSONB NOT NULL,
    error TEXT,
    status export_jobs_status NOT NULL,
    file_id TEXT,
    recipient_email TEXT NOT NULL,
    recipient_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE export_jobs ADD CONSTRAINT export_jobs_file_id_fkey
    FOREIGN KEY (file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE SET NULL;

DROP TABLE framework_exports;
