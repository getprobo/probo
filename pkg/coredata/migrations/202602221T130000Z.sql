CREATE TABLE compliance_badges (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL,
    icon_file_id TEXT NOT NULL REFERENCES files(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    rank INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT compliance_badges_trust_center_id_rank_key
        UNIQUE (trust_center_id, rank)
        DEFERRABLE INITIALLY DEFERRED
);
