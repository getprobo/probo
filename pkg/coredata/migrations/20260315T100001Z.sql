-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

-- Merge nonconformities and continual_improvements into findings

CREATE TYPE findings_kind AS ENUM ('NONCONFORMITY', 'OBSERVATION', 'EXCEPTION');
CREATE TYPE findings_status AS ENUM ('OPEN', 'IN_PROGRESS', 'CLOSED', 'RISK_ACCEPTED', 'MITIGATED', 'FALSE_POSITIVE');
CREATE TYPE findings_priority AS ENUM ('LOW', 'MEDIUM', 'HIGH');

CREATE TABLE findings (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    snapshot_id TEXT,
    source_id TEXT,
    kind findings_kind NOT NULL,
    reference_id TEXT NOT NULL,
    description TEXT,
    source TEXT,
    risk_id TEXT,
    identified_on DATE,
    root_cause TEXT,
    corrective_action TEXT,
    owner_id TEXT,
    due_date DATE,
    status findings_status NOT NULL,
    priority findings_priority NOT NULL,
    effectiveness_check TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT fk_findings_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id) ON DELETE CASCADE,

    CONSTRAINT fk_findings_owner
        FOREIGN KEY (owner_id)
        REFERENCES iam_membership_profiles(id),

    CONSTRAINT fk_findings_snapshot
        FOREIGN KEY (snapshot_id)
        REFERENCES snapshots(id) ON DELETE CASCADE,

    CONSTRAINT fk_findings_risk
        FOREIGN KEY (risk_id)
        REFERENCES risks(id) ON DELETE RESTRICT,

    UNIQUE (source_id, snapshot_id)
);

CREATE INDEX idx_findings_organization_snapshot ON findings (organization_id, snapshot_id);
CREATE UNIQUE INDEX idx_findings_org_reference_id ON findings (organization_id, reference_id) WHERE snapshot_id IS NULL;

-- Create findings_audits junction table
CREATE TABLE findings_audits (
    finding_id TEXT NOT NULL REFERENCES findings(id) ON DELETE CASCADE ON UPDATE CASCADE,
    audit_id TEXT NOT NULL REFERENCES audits(id) ON DELETE CASCADE ON UPDATE CASCADE,
    reference_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (finding_id, audit_id)
);

CREATE INDEX idx_findings_audits_audit_id ON findings_audits (audit_id);

-- Pre-generate IDs and FND-XXX reference_ids for live nonconformities
CREATE TEMP TABLE nc_id_map AS
SELECT
    nc.id AS old_id,
    generate_gid(decode_base64_unpadded(nc.tenant_id), 67) AS new_id,
    nc.organization_id,
    nc.reference_id AS old_reference_id
FROM nonconformities nc WHERE nc.snapshot_id IS NULL;

-- Assign FND-XXX reference_ids per organization across all live records
-- First, combine all live records from both tables with a stable ordering
CREATE TEMP TABLE all_live_findings AS
SELECT
    nc_id_map.old_id,
    nc_id_map.new_id,
    nc_id_map.organization_id,
    nc_id_map.old_reference_id,
    'NC' AS source_type,
    nonconformities.created_at
FROM nc_id_map
JOIN nonconformities ON nonconformities.id = nc_id_map.old_id;

-- Pre-generate IDs for live continual improvements
CREATE TEMP TABLE ci_id_map AS
SELECT
    ci.id AS old_id,
    generate_gid(decode_base64_unpadded(ci.tenant_id), 67) AS new_id,
    ci.organization_id,
    ci.reference_id AS old_reference_id
FROM continual_improvements ci WHERE ci.snapshot_id IS NULL;

INSERT INTO all_live_findings (old_id, new_id, organization_id, old_reference_id, source_type, created_at)
SELECT
    ci_id_map.old_id,
    ci_id_map.new_id,
    ci_id_map.organization_id,
    ci_id_map.old_reference_id,
    'CI',
    ci.created_at
FROM ci_id_map
JOIN continual_improvements ci ON ci.id = ci_id_map.old_id;

-- Generate sequential FND-XXX reference_ids per organization
CREATE TEMP TABLE finding_ref_map AS
SELECT
    old_id,
    new_id,
    organization_id,
    old_reference_id,
    source_type,
    'FND-' || LPAD(ROW_NUMBER() OVER (PARTITION BY organization_id ORDER BY created_at, old_id)::TEXT, 3, '0') AS new_reference_id
FROM all_live_findings;

-- Insert live nonconformities with FND-XXX reference_ids
INSERT INTO findings (
    id, tenant_id, organization_id, snapshot_id, source_id,
    kind, reference_id, description, source,
    identified_on, root_cause, corrective_action,
    owner_id, due_date, status, priority,
    effectiveness_check, created_at, updated_at
)
SELECT
    m.new_id,
    nc.tenant_id, nc.organization_id, NULL, NULL,
    'NONCONFORMITY'::findings_kind, m.new_reference_id, nc.description, NULL,
    nc.date_identified, nc.root_cause, nc.corrective_action,
    nc.owner_profile_id, nc.due_date, nc.status::text::findings_status, 'MEDIUM'::findings_priority,
    nc.effectiveness_check, nc.created_at, nc.updated_at
FROM nonconformities nc
JOIN finding_ref_map m ON nc.id = m.old_id AND m.source_type = 'NC';

-- Migrate nonconformity audit associations (old reference_id goes on junction table)
INSERT INTO findings_audits (finding_id, audit_id, reference_id, organization_id, tenant_id, created_at)
SELECT m.new_id, nc.audit_id, nc.reference_id, nc.organization_id, nc.tenant_id, nc.created_at
FROM nonconformities nc
JOIN finding_ref_map m ON nc.id = m.old_id AND m.source_type = 'NC'
WHERE nc.audit_id IS NOT NULL;

-- Insert snapshot nonconformities with source_id pointing to live finding
-- Snapshot findings keep the same reference_id as their live counterpart
INSERT INTO findings (
    id, tenant_id, organization_id, snapshot_id, source_id,
    kind, reference_id, description, source,
    identified_on, root_cause, corrective_action,
    owner_id, due_date, status, priority,
    effectiveness_check, created_at, updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(nc.tenant_id), 67),
    nc.tenant_id, nc.organization_id, nc.snapshot_id, m.new_id,
    'NONCONFORMITY'::findings_kind, m.new_reference_id, nc.description, NULL,
    nc.date_identified, nc.root_cause, nc.corrective_action,
    nc.owner_profile_id, nc.due_date, nc.status::text::findings_status, 'MEDIUM'::findings_priority,
    nc.effectiveness_check, nc.created_at, nc.updated_at
FROM nonconformities nc
LEFT JOIN finding_ref_map m ON nc.source_id = m.old_id AND m.source_type = 'NC'
WHERE nc.snapshot_id IS NOT NULL;

-- Migrate snapshot nonconformity audit associations
INSERT INTO findings_audits (finding_id, audit_id, reference_id, organization_id, tenant_id, created_at)
SELECT f.id, nc.audit_id, nc.reference_id, nc.organization_id, nc.tenant_id, nc.created_at
FROM nonconformities nc
JOIN findings f ON f.snapshot_id = nc.snapshot_id
    AND f.source_id IS NOT NULL
    AND f.kind = 'NONCONFORMITY'
    AND EXISTS (
        SELECT 1 FROM finding_ref_map m
        WHERE m.old_id = nc.source_id
        AND m.source_type = 'NC'
        AND f.source_id = m.new_id
    )
WHERE nc.snapshot_id IS NOT NULL AND nc.audit_id IS NOT NULL;

-- Insert live continual improvements with FND-XXX reference_ids
INSERT INTO findings (
    id, tenant_id, organization_id, snapshot_id, source_id,
    kind, reference_id, description, source,
    identified_on, root_cause, corrective_action,
    owner_id, due_date, status, priority,
    effectiveness_check, created_at, updated_at
)
SELECT
    m.new_id,
    ci.tenant_id, ci.organization_id, NULL, NULL,
    'OBSERVATION'::findings_kind, m.new_reference_id, ci.description, ci.source,
    NULL, NULL, NULL,
    ci.owner_profile_id, ci.target_date, ci.status::text::findings_status, ci.priority::text::findings_priority,
    NULL, ci.created_at, ci.updated_at
FROM continual_improvements ci
JOIN finding_ref_map m ON ci.id = m.old_id AND m.source_type = 'CI';

-- Insert snapshot continual improvements with source_id pointing to live finding
INSERT INTO findings (
    id, tenant_id, organization_id, snapshot_id, source_id,
    kind, reference_id, description, source,
    identified_on, root_cause, corrective_action,
    owner_id, due_date, status, priority,
    effectiveness_check, created_at, updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(ci.tenant_id), 67),
    ci.tenant_id, ci.organization_id, ci.snapshot_id, m.new_id,
    'OBSERVATION'::findings_kind, m.new_reference_id, ci.description, ci.source,
    NULL, NULL, NULL,
    ci.owner_profile_id, ci.target_date, ci.status::text::findings_status, ci.priority::text::findings_priority,
    NULL, ci.created_at, ci.updated_at
FROM continual_improvements ci
LEFT JOIN finding_ref_map m ON ci.source_id = m.old_id AND m.source_type = 'CI'
WHERE ci.snapshot_id IS NOT NULL;

DROP TABLE finding_ref_map;
DROP TABLE all_live_findings;
DROP TABLE nc_id_map;
DROP TABLE ci_id_map;

-- Update snapshot types
UPDATE snapshots
SET type = 'FINDINGS'
WHERE type IN ('NONCONFORMITIES', 'CONTINUAL_IMPROVEMENTS');

-- Drop old tables
DROP TABLE nonconformities;
DROP TABLE continual_improvements;
DROP TYPE nonconformities_status;
DROP TYPE continual_improvements_status;
DROP TYPE continual_improvement_registries_priority;
