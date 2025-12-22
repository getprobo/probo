INSERT INTO iam_identity_profiles (tenant_id, id, identity_id, membership_id, full_name, created_at, updated_at)
SELECT
    '',
    generate_gid('\x0000000000000000'::bytea, 51),
    i.id,
    NULL,
    COALESCE(i.fullname, ''),
    i.created_at,
    NOW()
FROM identities i
WHERE NOT EXISTS (
    SELECT 1 FROM iam_identity_profiles p 
    WHERE p.identity_id = i.id AND p.membership_id IS NULL
);

INSERT INTO iam_identity_profiles (tenant_id, id, identity_id, membership_id, full_name, created_at, updated_at)
SELECT
    m.tenant_id,
    generate_gid(decode_base64_unpadded(m.tenant_id), 51),
    m.identity_id,
    m.id,
    COALESCE(i.fullname, ''),
    m.created_at,
    NOW()
FROM iam_memberships m
JOIN identities i ON i.id = m.identity_id
WHERE NOT EXISTS (
    SELECT 1 FROM iam_identity_profiles p 
    WHERE p.membership_id = m.id
);

ALTER TABLE identities DROP COLUMN fullname;