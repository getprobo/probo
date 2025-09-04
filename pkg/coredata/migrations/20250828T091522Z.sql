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



