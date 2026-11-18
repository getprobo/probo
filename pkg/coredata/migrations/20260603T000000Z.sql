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

ALTER TABLE audits
    ADD COLUMN report_file_id TEXT REFERENCES files(id);

ALTER TABLE trust_center_document_accesses
    ADD COLUMN report_file_id TEXT REFERENCES files(id);

ALTER TABLE trust_center_document_accesses
    DROP CONSTRAINT trust_center_document_accesses_check;

/* 25 is for FileEntityType */
WITH report_files AS (
    SELECT
        r.id                       AS report_id,
        generate_gid(decode_base64_unpadded(r.tenant_id), 25) AS file_id,
        r.tenant_id,
        r.organization_id,
        'probod'                   AS bucket_name,
        r.mime_type,
        r.filename                 AS file_name,
        r.object_key::uuid         AS file_key,
        r.size                     AS file_size,
        r.created_at,
        r.updated_at
    FROM reports r
),
inserted_files AS (
    INSERT INTO files (
        id, tenant_id, organization_id, bucket_name,
        mime_type, file_name, file_key, file_size,
        visibility, created_at, updated_at
    )
    SELECT
        file_id, tenant_id, organization_id, bucket_name,
        mime_type, file_name, file_key, file_size,
        'PRIVATE', created_at, updated_at
    FROM report_files
)
SELECT rf.report_id, rf.file_id
INTO TEMP TABLE report_file_mapping
FROM report_files rf;

UPDATE audits a
SET report_file_id = rfm.file_id
FROM report_file_mapping rfm
WHERE a.report_id = rfm.report_id;

UPDATE trust_center_document_accesses tcda
SET report_file_id = rfm.file_id,
    report_id = NULL
FROM report_file_mapping rfm
WHERE tcda.report_id = rfm.report_id;

ALTER TABLE trust_center_document_accesses
    ADD CONSTRAINT trust_center_document_accesses_check CHECK (
        (document_id IS NOT NULL)::int + (report_id IS NOT NULL)::int + (trust_center_file_id IS NOT NULL)::int + (report_file_id IS NOT NULL)::int = 1
    );

ALTER TABLE trust_center_document_accesses
    ADD CONSTRAINT trust_center_document_accesses_trust_center_access_id_report_file_key
    UNIQUE (trust_center_access_id, report_file_id);
