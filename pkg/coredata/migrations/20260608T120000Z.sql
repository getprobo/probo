-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

-- ADD COLUMN ... DEFAULT 1 already backfills every existing row to level 1, so
-- only the sub-third-parties (first_level = false) need correcting to level 2.
-- Scoping the UPDATE to them leaves first-level rows entirely untouched.
ALTER TABLE third_parties ADD COLUMN level integer NOT NULL DEFAULT 1;
UPDATE third_parties SET level = 2 WHERE first_level = false;
ALTER TABLE third_parties ALTER COLUMN level DROP DEFAULT;

-- Keep first_level around instead of dropping it so the change is reversible and
-- old code paths keep working during a rolling deploy. New inserts no longer set
-- it, so give it a default to satisfy the NOT NULL constraint.
ALTER TABLE third_parties ALTER COLUMN first_level SET DEFAULT true;

ALTER TABLE third_parties ADD COLUMN parent_third_party_id text REFERENCES third_parties(id) ON DELETE CASCADE;

-- Pick each non-first-level child's primary parent: the parent it will be
-- re-parented to in place. Only first_level = false rows are eligible — a
-- first-level row stays a root (its links become copies below, never an
-- in-place move). Prefer a top-level (first_level = true) parent so the child
-- anchors to a real root, then fall back to the relationship that was created
-- first. IDs are text GIDs, so MIN() would order them lexicographically rather
-- than chronologically — sort by the junction row's created_at instead, with
-- the parent id as a deterministic tie-break.
CREATE TEMP TABLE _tp_primary_parent AS
SELECT DISTINCT ON (tpr.child_third_party_id)
    tpr.child_third_party_id,
    tpr.parent_third_party_id
FROM third_party_third_parties tpr
JOIN third_parties parent ON parent.id = tpr.parent_third_party_id
JOIN third_parties child ON child.id = tpr.child_third_party_id
WHERE child.first_level = false
ORDER BY
    tpr.child_third_party_id,
    parent.first_level DESC,
    tpr.created_at ASC,
    tpr.parent_third_party_id ASC;

-- Assign every existing child its primary parent in place. Updating the row in
-- place preserves its ID and therefore every dependent record (risk
-- assessments, services, contacts, measure links, …) — nothing is deleted or
-- re-keyed for the common one-parent case.
-- Restricted to first_level = false: only sub-third-parties may gain a parent;
-- a top-level (first_level = true) row must stay at level 1 with no parent even
-- if it appears in the junction table by accident.
UPDATE third_parties tp
SET parent_third_party_id = primary_parent.parent_third_party_id
FROM _tp_primary_parent AS primary_parent
WHERE tp.id = primary_parent.child_third_party_id
  AND tp.first_level = false;

-- Re-qualify the names of the reparented sub-third-parties to the hierarchy
-- convention used by the console/vetting ("base (root/.../parent)"), matching
-- the copies created below. Names are rebuilt from base names (trailing " (…)"
-- stripped) so already-qualified rows are normalized rather than double-suffixed.
WITH RECURSIVE tp_path AS (
    SELECT
        id,
        trim(regexp_replace(name, '\s*\([^)]*\)\s*$', '')) AS path
    FROM third_parties
    WHERE parent_third_party_id IS NULL
    UNION ALL
    SELECT
        c.id,
        p.path || '/' || trim(regexp_replace(c.name, '\s*\([^)]*\)\s*$', '')) AS path
    FROM third_parties c
    JOIN tp_path p ON p.id = c.parent_third_party_id
)
UPDATE third_parties tp
SET name = trim(regexp_replace(tp.name, '\s*\([^)]*\)\s*$', '')) || ' (' || pp.path || ')'
FROM tp_path pp
WHERE pp.id = tp.parent_third_party_id
  AND tp.first_level = false;

-- Every junction row that is NOT an in-place reparent becomes its own copy: the
-- extra parents of a non-first-level child, plus every link whose child is
-- first_level = true (the root is preserved and a fresh sub-third-party copy is
-- created under the parent). Generate a fresh GID for each (child, parent) pair.
-- generate_gid() and parse_tenant_id() are defined in migration 20250420T120000Z.
CREATE TEMP TABLE _tp_copy_map AS
SELECT
    tpr.child_third_party_id AS old_id,
    tpr.parent_third_party_id AS parent_id,
    generate_gid(parse_tenant_id(tp.tenant_id), 7) AS new_id
FROM third_party_third_parties tpr
JOIN third_parties tp ON tp.id = tpr.child_third_party_id
WHERE NOT EXISTS (
    SELECT 1
    FROM _tp_primary_parent pp
    WHERE pp.child_third_party_id = tpr.child_third_party_id
      AND pp.parent_third_party_id = tpr.parent_third_party_id
);

-- Resolve each third party's hierarchy-qualified path (root → self, base names
-- joined by "/"), mirroring the console/vetting naming convention. A copy under
-- parent P is named "<child base> (<path of P>)", e.g. "Google Workspace (Probo)".
-- The base name strips any trailing " (…)" suffix exactly like the app regex.
WITH RECURSIVE tp_path AS (
    SELECT
        id,
        level,
        trim(regexp_replace(name, '\s*\([^)]*\)\s*$', '')) AS path
    FROM third_parties
    WHERE parent_third_party_id IS NULL
    UNION ALL
    SELECT
        c.id,
        c.level,
        p.path || '/' || trim(regexp_replace(c.name, '\s*\([^)]*\)\s*$', '')) AS path
    FROM third_parties c
    JOIN tp_path p ON p.id = c.parent_third_party_id
)
INSERT INTO third_parties (
    id,
    tenant_id,
    organization_id,
    parent_third_party_id,
    common_third_party_id,
    name,
    description,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    countries,
    business_owner_profile_id,
    security_owner_profile_id,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    show_on_trust_center,
    first_level,
    level,
    created_at,
    updated_at
)
SELECT
    m.new_id,
    tp.tenant_id,
    tp.organization_id,
    -- The parent is always an existing row that needs no remapping.
    m.parent_id,
    -- A copy is a relationship/subprocessor record, not the canonical
    -- catalog-linked vendor. Leave common_third_party_id NULL so the original
    -- root stays the single third party resolved for a common catalog entry
    -- (e.g. cookie-banner tracker mapping by common_third_party_id).
    NULL,
    -- Hierarchy-qualified name: child base name suffixed with the parent's path.
    trim(regexp_replace(tp.name, '\s*\([^)]*\)\s*$', '')) || ' (' || pp.path || ')',
    tp.description,
    tp.category,
    tp.headquarter_address,
    tp.legal_name,
    tp.website_url,
    tp.privacy_policy_url,
    tp.service_level_agreement_url,
    tp.data_processing_agreement_url,
    tp.business_associate_agreement_url,
    tp.subprocessors_list_url,
    tp.certifications,
    tp.countries,
    tp.business_owner_profile_id,
    tp.security_owner_profile_id,
    tp.status_page_url,
    tp.terms_of_service_url,
    tp.security_page_url,
    tp.trust_page_url,
    tp.show_on_trust_center,
    -- Copies are sub-third-parties (level >= 2), never first-level roots.
    false,
    -- Level follows the parent, not the copied child: a first-level child
    -- (level 1) copied under Probo (level 1) becomes a level-2 sub-third-party.
    pp.level + 1,
    tp.created_at,
    tp.updated_at
FROM _tp_copy_map m
JOIN third_parties tp ON tp.id = m.old_id
JOIN tp_path pp ON pp.id = m.parent_id;

INSERT INTO third_party_services (
    tenant_id,
    id,
    organization_id,
    third_party_id,
    name,
    description,
    created_at,
    updated_at
)
SELECT
    tp.tenant_id,
    generate_gid(parse_tenant_id(tp.tenant_id), 30),
    s.organization_id,
    m.new_id,
    s.name,
    s.description,
    s.created_at,
    s.updated_at
FROM _tp_copy_map m
JOIN third_parties tp ON tp.id = m.old_id
JOIN third_party_services s ON s.third_party_id = m.old_id;

INSERT INTO third_party_contacts (
    tenant_id,
    id,
    organization_id,
    third_party_id,
    full_name,
    email,
    phone,
    role,
    created_at,
    updated_at
)
SELECT
    tp.tenant_id,
    generate_gid(parse_tenant_id(tp.tenant_id), 26),
    c.organization_id,
    m.new_id,
    c.full_name,
    c.email,
    c.phone,
    c.role,
    c.created_at,
    c.updated_at
FROM _tp_copy_map m
JOIN third_parties tp ON tp.id = m.old_id
JOIN third_party_contacts c ON c.third_party_id = m.old_id;

DROP TABLE _tp_copy_map;
DROP TABLE _tp_primary_parent;
DROP TABLE third_party_third_parties;
