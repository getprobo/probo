CREATE TABLE trust_center_references (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    website_url TEXT NOT NULL,
    logo_file_id TEXT NOT NULL REFERENCES files(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
