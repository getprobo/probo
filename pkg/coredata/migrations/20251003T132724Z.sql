ALTER TABLE evidences
    ADD COLUMN evidence_file_id text;

/* 25 is for FileEntityType */
WITH
    evidence_files AS (
        SELECT
            e.id as evidence_id,
            generate_gid(decode_base64_unpadded(e.tenant_id), 25) as file_id,
            e.tenant_id,
            'probod' as bucket_name,
            e.mime_type,
            e.filename,
            e.object_key,
            e.size,
            e.created_at,
            e.updated_at
        FROM evidences e
        WHERE  e.object_key IS NOT NULL AND e.object_key != ''
    ),
    inserted_files AS (
        INSERT INTO files (id, tenant_id, bucket_name, mime_type, file_name, file_key, file_size, created_at, updated_at)
            SELECT file_id, tenant_id, bucket_name, mime_type, filename, object_key::uuid, size, created_at, updated_at
            FROM evidence_files
            RETURNING id, tenant_id
    )

SELECT ef.evidence_id, ef.file_id
INTO TEMP TABLE file_evidence_mapping
FROM evidence_files ef;

UPDATE evidences
SET evidence_file_id = fm.file_id
FROM file_evidence_mapping fm
WHERE evidences.id = fm.evidence_id;


ALTER TABLE evidences
    ALTER COLUMN filename DROP NOT NULL,
    ALTER COLUMN mime_type DROP NOT NULL,
    ALTER COLUMN size DROP NOT NULL,
    ALTER COLUMN object_key DROP NOT NULL;
