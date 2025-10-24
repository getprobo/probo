CREATE TABLE trust_center_files (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    file_id TEXT NOT NULL REFERENCES files(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    trust_center_visibility trust_center_visibility NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

ALTER TABLE trust_center_document_accesses ADD COLUMN trust_center_file_id TEXT REFERENCES trust_center_files(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE trust_center_document_accesses DROP CONSTRAINT trust_center_document_accesses_check;

ALTER TABLE trust_center_document_accesses ADD CONSTRAINT trust_center_document_accesses_check CHECK (
    (document_id IS NOT NULL)::int + (report_id IS NOT NULL)::int + (trust_center_file_id IS NOT NULL)::int = 1
);

ALTER TABLE trust_center_document_accesses ADD CONSTRAINT trust_center_document_accesses_trust_center_file_id_key UNIQUE (trust_center_access_id, trust_center_file_id);
