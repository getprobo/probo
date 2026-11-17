-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

CREATE TABLE search_resources (
    resource_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations (id) ON UPDATE CASCADE ON DELETE CASCADE,
    entity_type INT NOT NULL,
    title TEXT NOT NULL,
    body_vector tsvector NOT NULL,
    tag_vector tsvector NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (body_vector || tag_vector) STORED
);

CREATE INDEX search_resources_search_idx ON search_resources USING gin (search_vector);
CREATE INDEX search_resources_organization_id_idx ON search_resources (organization_id);

-- Backfill live rows.

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.version, '') || ' ' || COALESCE(r.source, '') || ' ' || COALESCE(r.purpose, '') || ' ' || COALESCE(r.intended_use_cases, '') || ' ' || COALESCE(r.autonomy_level, '') || ' ' || COALESCE(r.human_oversight_mechanism, '') || ' ' || COALESCE(r.key_stakeholders, '') || ' ' || COALESCE(r.data_sources_and_type, '') || ' ' || COALESCE(r.notes, ''), '-+', '  ')),
    ''::tsvector
FROM ai_systems r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.data_types_stored, ''), '-+', '  ')),
    ''::tsvector
FROM assets r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.firm, ''), '-+', '  ')),
    ''::tsvector
FROM audits r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.reference_id, '') || ' ' || COALESCE(r.impact_tolerance, '') || ' ' || COALESCE(r.notes, ''), '-+', '  ')),
    ''::tsvector
FROM business_functions r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.section_title, '') || ' ' || COALESCE(r.name, '') || ' ' || COALESCE(r.description, ''), '-+', '  ')),
    ''::tsvector
FROM controls r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, ''), '-+', '  ')),
    ''::tsvector
FROM data r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.reference_id, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.reference_id, '') || ' ' || COALESCE(r.description, '') || ' ' || COALESCE(r.url, ''), '-+', '  ')),
    ''::tsvector
FROM evidences r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.reference_id, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.reference_id, '') || ' ' || COALESCE(r.description, '') || ' ' || COALESCE(r.source, '') || ' ' || COALESCE(r.root_cause, '') || ' ' || COALESCE(r.corrective_action, '') || ' ' || COALESCE(r.effectiveness_check, ''), '-+', '  ')),
    ''::tsvector
FROM findings r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.description, ''), '-+', '  ')),
    ''::tsvector
FROM frameworks r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.code, '') || ' ' || COALESCE(r.operating_event, '') || ' ' || COALESCE(r.description, ''), '-+', '  ')),
    ''::tsvector
FROM internal_controls r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.requirement, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.requirement, '') || ' ' || COALESCE(r.area, '') || ' ' || COALESCE(r.source, '') || ' ' || COALESCE(r.regulator, '') || ' ' || COALESCE(r.actions_to_be_implemented, ''), '-+', '  ')),
    ''::tsvector
FROM obligations r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.purpose, '') || ' ' || COALESCE(r.data_subject_category, '') || ' ' || COALESCE(r.personal_data_category, '') || ' ' || COALESCE(r.recipients, '') || ' ' || COALESCE(r.location, '') || ' ' || COALESCE(r.retention_period, '') || ' ' || COALESCE(r.security_measures, ''), '-+', '  ')),
    ''::tsvector
FROM processing_activities r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.data_subject, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.data_subject, '') || ' ' || COALESCE(r.contact, '') || ' ' || COALESCE(r.details, '') || ' ' || COALESCE(r.action_taken, ''), '-+', '  ')),
    ''::tsvector
FROM rights_requests r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.reference_id, '') || ' ' || COALESCE(r.name, ''), '-+', '  ')),
    ''::tsvector
FROM risks r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.description::text, ''), '-+', '  ')),
    ''::tsvector
FROM risk_analyses r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, ''), '-+', '  ')),
    ''::tsvector
FROM risk_analysis_boundaries r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, ''), '-+', '  ')),
    ''::tsvector
FROM risk_analysis_diagrams r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, ''), '-+', '  ')),
    ''::tsvector
FROM risk_analysis_processes r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.description, ''), '-+', '  ')),
    ''::tsvector
FROM risk_analysis_scenarios r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.category, ''), '-+', '  ')),
    ''::tsvector
FROM risk_analysis_threats r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, ''), '-+', '  ')),
    ''::tsvector
FROM statements_of_applicability r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.reference_id, '') || ' ' || COALESCE(r.content::text, ''), '-+', '  ')),
    ''::tsvector
FROM tasks r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.name, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.name, '') || ' ' || COALESCE(r.description, '') || ' ' || COALESCE(r.legal_name, ''), '-+', '  ')),
    ''::tsvector
FROM third_parties r;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.title, '')), ''), r.id),
    to_tsvector('simple', translate(COALESCE(r.title, ''), '-+', '  ')),
    ''::tsvector
FROM documents r
WHERE r.deleted_at IS NULL;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(r.hostname, r.serial_number, r.hardware_uuid, '')), ''), r.id),
    to_tsvector(
        'simple',
        translate(
            COALESCE(r.hostname, '') || ' ' ||
            COALESCE(r.serial_number, '') || ' ' ||
            COALESCE(r.hardware_uuid, '') || ' ' ||
            COALESCE(r.os_version, '') || ' ' ||
            COALESCE(r.agent_version, '') || ' ' ||
            COALESCE(r.platform::text, '') || ' ' ||
            COALESCE(r.labels::text, ''),
            '-+', '  '
        )
    ),
    ''::tsvector
FROM devices r
WHERE r.deleted_at IS NULL;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(pa.name, '')), ''), r.id),
    to_tsvector(
        'simple',
        translate(
            COALESCE(pa.name, '') || ' ' ||
            COALESCE(r.description, '') || ' ' ||
            COALESCE(r.necessity_and_proportionality, '') || ' ' ||
            COALESCE(r.potential_risk, '') || ' ' ||
            COALESCE(r.mitigations, ''),
            '-+', '  '
        )
    ),
    ''::tsvector
FROM processing_activity_data_protection_impact_assessments r
LEFT JOIN processing_activities pa ON pa.id = r.processing_activity_id;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(pa.name, '')), ''), r.id),
    to_tsvector(
        'simple',
        translate(
            COALESCE(pa.name, '') || ' ' ||
            COALESCE(r.data_subjects, '') || ' ' ||
            COALESCE(r.legal_mechanism, '') || ' ' ||
            COALESCE(r.transfer, '') || ' ' ||
            COALESCE(r.local_law_risk, '') || ' ' ||
            COALESCE(r.supplementary_measures, ''),
            '-+', '  '
        )
    ),
    ''::tsvector
FROM processing_activity_transfer_impact_assessments r
LEFT JOIN processing_activities pa ON pa.id = r.processing_activity_id;

INSERT INTO search_resources (
    resource_id,
    tenant_id,
    organization_id,
    entity_type,
    title,
    body_vector,
    tag_vector
)
SELECT
    r.id,
    r.tenant_id,
    r.organization_id,
    extract_entity_type(parse_gid(r.id)),
    COALESCE(NULLIF(BTRIM(COALESCE(risk.name, '')), ''), r.id),
    to_tsvector(
        'simple',
        translate(
            COALESCE(risk.name, '') || ' ' || COALESCE(r.treatment::text, ''),
            '-+', '  '
        )
    ),
    ''::tsvector
FROM treatment_plans r
LEFT JOIN risks risk ON risk.id = r.risk_id;

UPDATE search_resources sr
SET tag_vector = COALESCE(tag_data.vector, ''::tsvector)
FROM (
    SELECT
        ta.resource_id,
        to_tsvector(
            'simple',
            translate(string_agg(t.name, ' '), '-+', '  ')
        ) AS vector
    FROM tag_assignments ta
    INNER JOIN tags t ON t.id = ta.tag_id
    GROUP BY ta.resource_id
) tag_data
WHERE sr.resource_id = tag_data.resource_id;
