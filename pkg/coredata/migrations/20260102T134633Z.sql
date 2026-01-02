CREATE TYPE state_of_applicability_control_state AS ENUM (
    'EXCLUDED',
    'IMPLEMENTED',
    'NOT_IMPLEMENTED'
);

CREATE TABLE states_of_applicability (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    source_id TEXT,
    snapshot_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT state_of_applicabilities_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT state_of_applicabilities_snapshot_id_fkey
        FOREIGN KEY (snapshot_id)
        REFERENCES snapshots(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX states_of_applicability_source_id_snapshot_id_uniq
    ON states_of_applicability (source_id, snapshot_id)
    WHERE snapshot_id IS NULL;

CREATE UNIQUE INDEX states_of_applicability_name_organization_id_uniq
    ON states_of_applicability (name, organization_id)
    WHERE snapshot_id IS NULL;

CREATE TABLE states_of_applicability_controls (
    state_of_applicability_id TEXT NOT NULL REFERENCES states_of_applicability(id) ON DELETE CASCADE ON UPDATE CASCADE,
    control_id TEXT NOT NULL REFERENCES controls(id) ON DELETE CASCADE ON UPDATE CASCADE,
    tenant_id TEXT NOT NULL,
    snapshot_id TEXT,
    state state_of_applicability_control_state NOT NULL,
    exclusion_justification TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (state_of_applicability_id, control_id),

    CONSTRAINT states_of_applicability_controls_snapshot_id_fkey
        FOREIGN KEY (snapshot_id)
        REFERENCES snapshots(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

INSERT INTO states_of_applicability (
    id,
    tenant_id,
    organization_id,
    name,
    description,
    source_id,
    snapshot_id,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(f.tenant_id), 48) as id,
    f.tenant_id,
    f.organization_id,
    f.name,
    NULL as description,
    NULL as source_id,
    NULL as snapshot_id,
    NOW() as created_at,
    NOW() as updated_at
FROM frameworks f
WHERE NOT EXISTS (
    SELECT 1
    FROM states_of_applicability soa
    WHERE soa.name = f.name
    AND soa.snapshot_id IS NULL
);

INSERT INTO states_of_applicability_controls (
    state_of_applicability_id,
    control_id,
    tenant_id,
    snapshot_id,
    state,
    exclusion_justification,
    created_at
)
SELECT
    soa.id as state_of_applicability_id,
    c.id as control_id,
    c.tenant_id,
    NULL as snapshot_id,
    CASE
        WHEN c.status = 'EXCLUDED' THEN 'EXCLUDED'::state_of_applicability_control_state
        ELSE 'IMPLEMENTED'::state_of_applicability_control_state
    END as state,
    c.exclusion_justification,
    NOW() as created_at
FROM frameworks f
JOIN states_of_applicability soa ON soa.name = f.name AND soa.snapshot_id IS NULL
JOIN controls c ON c.framework_id = f.id
WHERE NOT EXISTS (
    SELECT 1
    FROM states_of_applicability_controls soac
    WHERE soac.state_of_applicability_id = soa.id
    AND soac.control_id = c.id
);
