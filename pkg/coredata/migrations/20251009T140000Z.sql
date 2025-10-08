ALTER TABLE organizations ALTER COLUMN logo_object_key SET DEFAULT '';

ALTER TABLE organizations
    ADD COLUMN horizontal_logo_file_id TEXT,
    ADD CONSTRAINT organizations_horizontal_logo_file_id_fkey
        FOREIGN KEY (horizontal_logo_file_id)
        REFERENCES files(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT;

ALTER TABLE organizations
    ADD COLUMN logo_file_id TEXT;

/* 25 is for FileEntityType */
WITH
    logo_files AS (
        SELECT
            o.id as organization_id,
            generate_gid(decode_base64_unpadded(o.tenant_id), 25) as file_id,
            o.tenant_id,
            'probod' as bucket_name,
            'image/png' as mime_type,
            'logo.png' as file_name,
            o.logo_object_key,
            o.created_at,
            o.updated_at
        FROM organizations o
        WHERE o.logo_object_key IS NOT NULL AND o.logo_object_key != ''
    ),
    inserted_files AS (
        INSERT INTO files (id, tenant_id, bucket_name, mime_type, file_name, file_key, file_size, created_at, updated_at)
            SELECT file_id, tenant_id, bucket_name, mime_type, file_name, logo_object_key::uuid, 0, created_at, updated_at
            FROM logo_files
            RETURNING id, tenant_id
    )

SELECT lf.organization_id, lf.file_id
INTO TEMP TABLE file_logo_mapping
FROM logo_files lf;

UPDATE organizations
SET logo_file_id = fm.file_id
FROM file_logo_mapping fm
WHERE organizations.id = fm.organization_id;

ALTER TABLE organizations
    ADD CONSTRAINT organizations_logo_file_id_fkey
        FOREIGN KEY (logo_file_id)
        REFERENCES files(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT;
