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

ALTER TABLE generated_documents
    ADD COLUMN risks_document_id TEXT REFERENCES documents(id) ON DELETE SET NULL;

-- Backfill controls_documents from the legacy controls_snapshots links.
-- For every snapshot type whose register is now an org-level generated
-- document, link each control that was attached to a snapshot to the
-- corresponding generated document. Best effort: skips rows whose target
-- document hasn't been created yet (the matching `cmd/migrate-*` data
-- migration must have already run). ON CONFLICT keeps the migration
-- idempotent and tolerant of pre-existing mappings.
INSERT INTO controls_documents (control_id, document_id, organization_id, tenant_id, created_at)
SELECT DISTINCT
    cs.control_id,
    CASE s.type
        WHEN 'RISKS'                 THEN gd.risks_document_id
        WHEN 'VENDORS'               THEN gd.vendors_document_id
        WHEN 'ASSETS'                THEN gd.asset_list_document_id
        WHEN 'DATA'                  THEN gd.data_document_id
        WHEN 'FINDINGS'              THEN gd.findings_document_id
        WHEN 'OBLIGATIONS'           THEN gd.obligations_document_id
        WHEN 'PROCESSING_ACTIVITIES' THEN gd.processing_activities_document_id
    END AS document_id,
    s.organization_id,
    s.tenant_id,
    NOW()
FROM controls_snapshots cs
INNER JOIN snapshots s ON s.id = cs.snapshot_id
LEFT JOIN generated_documents gd ON gd.organization_id = s.organization_id
WHERE s.type IN (
        'RISKS',
        'VENDORS',
        'ASSETS',
        'DATA',
        'FINDINGS',
        'OBLIGATIONS',
        'PROCESSING_ACTIVITIES'
    )
    AND CASE s.type
        WHEN 'RISKS'                 THEN gd.risks_document_id
        WHEN 'VENDORS'               THEN gd.vendors_document_id
        WHEN 'ASSETS'                THEN gd.asset_list_document_id
        WHEN 'DATA'                  THEN gd.data_document_id
        WHEN 'FINDINGS'              THEN gd.findings_document_id
        WHEN 'OBLIGATIONS'           THEN gd.obligations_document_id
        WHEN 'PROCESSING_ACTIVITIES' THEN gd.processing_activities_document_id
    END IS NOT NULL
ON CONFLICT DO NOTHING;

-- For STATEMENTS_OF_APPLICABILITY snapshots, the published document lives on
-- the source SOA (the live row, snapshot_id IS NULL). Link controls that
-- were attached to a SOA snapshot to that source SOA's document.
INSERT INTO controls_documents (control_id, document_id, organization_id, tenant_id, created_at)
SELECT DISTINCT
    cs.control_id,
    live_soa.document_id,
    s.organization_id,
    s.tenant_id,
    NOW()
FROM controls_snapshots cs
INNER JOIN snapshots s ON s.id = cs.snapshot_id
INNER JOIN statements_of_applicability snap_soa ON snap_soa.snapshot_id = s.id
INNER JOIN statements_of_applicability live_soa
    ON live_soa.id = snap_soa.source_id
    AND live_soa.snapshot_id IS NULL
WHERE s.type = 'STATEMENTS_OF_APPLICABILITY'
    AND live_soa.document_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Drop the trailing " List" suffix from previously published register
-- documents so the version title matches the new naming convention used by
-- the publish flow. Restricted to REGISTER document types so unrelated
-- documents that happen to share a title aren't touched.
UPDATE document_versions SET title = 'Assets'      WHERE title = 'Asset List'      AND document_type = 'REGISTER';
UPDATE document_versions SET title = 'Data'        WHERE title = 'Data List'       AND document_type = 'REGISTER';
UPDATE document_versions SET title = 'Findings'    WHERE title = 'Finding List'    AND document_type = 'REGISTER';
UPDATE document_versions SET title = 'Obligations' WHERE title = 'Obligation List' AND document_type = 'REGISTER';
