CREATE TABLE compliance_external_urls (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    rank INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

ALTER TABLE compliance_external_urls
ADD CONSTRAINT compliance_external_urls_trust_center_id_rank_key
    UNIQUE (trust_center_id, rank)
    DEFERRABLE INITIALLY DEFERRED;
