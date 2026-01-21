CREATE TABLE states_of_applicability (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    name TEXT NOT NULL,
    source_id TEXT,
    snapshot_id TEXT,
    owner_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT states_of_applicability_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT states_of_applicability_snapshot_id_fkey
        FOREIGN KEY (snapshot_id)
        REFERENCES snapshots(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT states_of_applicability_owner_id_fkey
        FOREIGN KEY (owner_id)
        REFERENCES peoples(id)
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
    id TEXT PRIMARY KEY,
    state_of_applicability_id TEXT NOT NULL REFERENCES states_of_applicability(id) ON DELETE CASCADE ON UPDATE CASCADE,
    control_id TEXT NOT NULL REFERENCES controls(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    organization_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    snapshot_id TEXT,
    applicability BOOLEAN NOT NULL,
    justification TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT states_of_applicability_controls_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT states_of_applicability_controls_snapshot_id_fkey
        FOREIGN KEY (snapshot_id)
        REFERENCES snapshots(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    UNIQUE (state_of_applicability_id, control_id)
);

ALTER TABLE controls ADD COLUMN best_practice BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE controls ALTER COLUMN best_practice DROP DEFAULT;

ALTER TYPE snapshots_type ADD VALUE 'STATES_OF_APPLICABILITY';

CREATE TYPE obligation_type AS ENUM (
    'LEGAL',
    'CONTRACTUAL'
);

ALTER TABLE obligations ADD COLUMN type obligation_type NOT NULL DEFAULT 'LEGAL';
ALTER TABLE obligations ALTER COLUMN type DROP DEFAULT;

CREATE TABLE controls_obligations (
    control_id TEXT NOT NULL REFERENCES controls(id) ON DELETE CASCADE ON UPDATE CASCADE,
    obligation_id TEXT NOT NULL REFERENCES obligations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (control_id, obligation_id)
);

INSERT INTO states_of_applicability (
    id,
    tenant_id,
    organization_id,
    name,
    source_id,
    snapshot_id,
    owner_id,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(f.tenant_id), 49) as id,
    f.tenant_id,
    f.organization_id,
    f.name,
    NULL as source_id,
    NULL as snapshot_id,
    (SELECT id FROM peoples WHERE tenant_id = f.tenant_id LIMIT 1) as owner_id,
    NOW() as created_at,
    NOW() as updated_at
FROM frameworks f
WHERE NOT EXISTS (
    SELECT 1
    FROM states_of_applicability soa
    WHERE soa.name = f.name
    AND soa.tenant_id = f.tenant_id
    AND soa.snapshot_id IS NULL
)
AND EXISTS (
    SELECT 1 FROM peoples WHERE tenant_id = f.tenant_id
)
-- We only use exclude with ISO 27001 FIXME
AND EXISTS (
    SELECT 1
    FROM controls c
    WHERE c.framework_id = f.id
    AND c.tenant_id = f.tenant_id
    AND (
        c.status = 'EXCLUDED'
        OR (c.exclusion_justification IS NOT NULL AND c.exclusion_justification != '')
    )
)
AND organization_id = 'e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg';

INSERT INTO states_of_applicability_controls (
    id,
    state_of_applicability_id,
    control_id,
    organization_id,
    tenant_id,
    snapshot_id,
    applicability,
    justification,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(c.tenant_id), 50) as id,
    soa.id as state_of_applicability_id,
    c.id as control_id,
    c.organization_id,
    c.tenant_id,
    NULL as snapshot_id,
    CASE
        WHEN c.status = 'EXCLUDED' THEN FALSE
        ELSE TRUE
    END as applicability,
    c.exclusion_justification,
    NOW() as created_at,
    NOW() as updated_at
FROM frameworks f
JOIN states_of_applicability soa ON soa.name = f.name AND soa.snapshot_id IS NULL AND soa.tenant_id = f.tenant_id
JOIN controls c ON c.framework_id = f.id
WHERE NOT EXISTS (
    SELECT 1
    FROM states_of_applicability_controls soac
    WHERE soac.state_of_applicability_id = soa.id
    AND soac.control_id = c.id
)
AND f.id = 'e5IaD7ibAAEAAQAAAZsNt6Js2dJrgpJG'
AND f.organization_id = 'e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg';



WITH todelete AS (
SELECT soac.id
FROM states_of_applicability soa
JOIN states_of_applicability_controls soac ON soac.state_of_applicability_id = soa.id
WHERE soa.tenant_id != soac.tenant_id
) 
DELETE FROM states_of_applicability_controls WHERE id IN (SELECT id FROM todelete)
