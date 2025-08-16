CREATE TYPE nonconformity_registries_status AS ENUM (
    'OPEN',
    'IN_PROGRESS',
    'CLOSED'
);

CREATE TABLE nonconformity_registries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    description TEXT,
    audit_id TEXT NOT NULL,
    date_identified DATE,
    root_cause TEXT NOT NULL,
    corrective_action TEXT,
    owner_id TEXT NOT NULL,
    due_date DATE,
    status nonconformity_registries_status NOT NULL,
    effectiveness_check TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT nonconformity_registries_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT nonconformity_registries_owner_id_fkey
        FOREIGN KEY (owner_id)
        REFERENCES peoples(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    CONSTRAINT nonconformity_registries_audit_id_fkey
        FOREIGN KEY (audit_id)
        REFERENCES audits(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);
