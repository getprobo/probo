ALTER TABLE
    vendor_compliance_reports
    ADD COLUMN report_file_id text
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;

WITH
    /* 25 is for FileEntityType */
    vcr_files AS (
        SELECT
            vcr.id as report_id,
            generate_gid(decode_base64_unpadded(vcr.tenant_id), 25) as file_id,
            vcr.tenant_id,
            'probod' as bucket_name,
            'application/pdf' as mime_type,
            vcr.report_name,
            vcr.file_key,
            vcr.file_size,
            vcr.created_at,
            vcr.updated_at
        FROM vendor_compliance_reports vcr
        WHERE snapshot_id IS NULL
    ),
    inserted_files AS (
        INSERT INTO files (id, tenant_id, bucket_name, mime_type, file_name, file_key, file_size, created_at, updated_at)
            SELECT file_id, tenant_id, bucket_name, mime_type, report_name, file_key::uuid, file_size, created_at, updated_at
            FROM vcr_files
            RETURNING id, tenant_id
    )
SELECT vf.report_id, vf.file_id
INTO TEMP TABLE file_vcr_mapping
FROM vcr_files vf;

UPDATE vendor_compliance_reports
SET report_file_id = fv.file_id
FROM file_vcr_mapping fv
WHERE vendor_compliance_reports.id = fv.report_id;

ALTER TABLE vendor_compliance_reports
    ALTER COLUMN file_key DROP NOT NULL,
    ALTER COLUMN file_size DROP NOT NULL;

UPDATE vendor_compliance_reports
SET report_file_id = f.id
FROM files f
WHERE vendor_compliance_reports.snapshot_id IS NOT NULL
    AND vendor_compliance_reports.report_file_id IS NULL
    AND vendor_compliance_reports.file_key IS NOT NULL
    AND f.file_key = vendor_compliance_reports.file_key::uuid
    AND f.tenant_id = vendor_compliance_reports.tenant_id;
