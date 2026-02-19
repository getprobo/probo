-- =============================================================================
-- ⚠️  BREAKING CHANGE (intentional, accepted for simplicity):
--   Several columns and tables are renamed, which will break any code or
--   queries still referencing the old names. Deploy the matching code changes
--   atomically with this migration.
--
-- Strategy:
--   - No data is deleted.
--   - The old reports table (file metadata) is renamed to deprecated_reports.
--   - audits.report_id is renamed to deprecated_report_id.
--   - trust_center_document_accesses.report_id is renamed to deprecated_report_id
--     so the original values are preserved.
--   - A new report_id column is added that holds proper report entity IDs.
-- =============================================================================

-- Part 1: Add file_id to audits, referencing the files table
ALTER TABLE audits ADD COLUMN file_id TEXT REFERENCES files(id) ON DELETE SET NULL;

-- Part 2: Migrate old reports (file metadata rows) into the files table
--         and build a temporary mapping old_report_id → new_file_id
WITH
    /* 25 = FileEntityType */
    report_files AS (
        SELECT
            r.id                                                        AS report_id,
            generate_gid(decode_base64_unpadded(r.tenant_id), 25)      AS file_id,
            r.tenant_id,
            r.organization_id,
            'probod' as bucket_name,
            r.mime_type,
            r.filename                                                  AS file_name,
            r.object_key                                                AS file_key,
            r.size                                                      AS file_size,
            r.created_at,
            r.updated_at
        FROM reports r
    ),
    inserted_files AS (
        INSERT INTO files (id, tenant_id, organization_id, bucket_name, mime_type, file_name, file_key, file_size, created_at, updated_at)
            SELECT file_id, tenant_id, organization_id, bucket_name, mime_type, file_name, file_key::uuid, file_size, created_at, updated_at
            FROM report_files
            RETURNING id, tenant_id
    )
SELECT rf.report_id, rf.file_id
INTO TEMP TABLE report_file_mapping
FROM report_files rf;

-- Part 3: Populate audits.file_id from the mapping
UPDATE audits a
SET file_id = m.file_id
FROM report_file_mapping m
WHERE a.report_id = m.report_id;

-- Part 4: Rename audits.report_id → deprecated_report_id (data preserved, not deleted)
ALTER TABLE audits RENAME COLUMN report_id TO deprecated_report_id;

-- Part 5: Rename old reports table → deprecated_reports (data preserved, not deleted)
ALTER TABLE reports RENAME TO deprecated_reports;

-- Part 6: Rename audits → reports and align related tables
ALTER TABLE audits RENAME TO reports;
ALTER TABLE controls_audits RENAME TO controls_reports;
ALTER TABLE controls_reports RENAME COLUMN audit_id TO report_id;
ALTER TABLE nonconformities RENAME COLUMN audit_id TO report_id;

-- Part 7: Rename trust_center_document_accesses.report_id → deprecated_report_id
--         (preserves the old values; FK automatically retargets to deprecated_reports)
ALTER TABLE trust_center_document_accesses
    RENAME COLUMN report_id TO deprecated_report_id;

-- Part 8: Add the new report_id column (will hold proper reports entity IDs)
ALTER TABLE trust_center_document_accesses
    ADD COLUMN report_id TEXT;

-- Part 9: Populate report_id by joining the mapping to the new reports table
--         mapping: deprecated_report_id → file_id → reports.file_id → reports.id
UPDATE trust_center_document_accesses tcda
SET report_id = r.id
FROM report_file_mapping m
JOIN reports r ON r.file_id = m.file_id
WHERE m.report_id = tcda.deprecated_report_id
    AND tcda.deprecated_report_id IS NOT NULL;

DROP TABLE IF EXISTS report_file_mapping;

-- Part 10: Rename the old FK (still named after report_id) to match its new column,
--          then add a fresh FK on the new report_id column
ALTER TABLE trust_center_document_accesses
    RENAME CONSTRAINT trust_center_document_accesses_report_id_fkey
    TO trust_center_document_accesses_deprecated_report_id_fkey;

ALTER TABLE trust_center_document_accesses
    ADD CONSTRAINT trust_center_document_accesses_report_id_fkey
    FOREIGN KEY (report_id) REFERENCES reports(id) ON UPDATE CASCADE ON DELETE CASCADE;

-- Part 11: Update the check constraint to use the new report_id column
--          (renaming deprecated_report_id automatically updated the constraint to reference
--          deprecated_report_id, but new inserts set report_id, not deprecated_report_id)
ALTER TABLE trust_center_document_accesses
    DROP CONSTRAINT trust_center_document_accesses_check;

ALTER TABLE trust_center_document_accesses
    ADD CONSTRAINT trust_center_document_accesses_check CHECK (
        (document_id IS NOT NULL)::int + (report_id IS NOT NULL)::int + (trust_center_file_id IS NOT NULL)::int = 1
    );

-- =============================================================================
-- TODO: run the following in a future migration once all deployments use new code
--
-- 1. ALTER TABLE trust_center_document_accesses
--        DROP CONSTRAINT trust_center_document_accesses_deprecated_report_id_fkey
-- 2. ALTER TABLE trust_center_document_accesses DROP COLUMN deprecated_report_id
-- 3. ALTER TABLE reports DROP COLUMN deprecated_report_id
-- 4. DROP TABLE deprecated_reports
-- =============================================================================
