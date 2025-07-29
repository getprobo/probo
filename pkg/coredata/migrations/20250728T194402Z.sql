CREATE TABLE trust_center_accesses (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id) ON DELETE CASCADE,
    email CITEXT NOT NULL
    name TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(trust_center_id, email)
);
