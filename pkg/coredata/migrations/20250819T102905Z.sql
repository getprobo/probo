CREATE TYPE snapshots_type AS ENUM (
    'RISKS',
    'VENDORS',
    'ASSETS',
    'DATA',
    'NON_CONFORMITY_REGISTRIES',
    'COMPLIANCE_REGISTRIES'
);

CREATE TABLE snapshots (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    type snapshots_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT snapshots_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);
