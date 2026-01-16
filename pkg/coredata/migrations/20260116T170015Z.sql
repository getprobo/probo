WITH tas AS (
    SELECT
        name,
        email
    FROM
        trust_center_accesses
)
INSERT INTO
    identities (
        id,
        created_at,
        updated_at,
        email_address,
        email_address_verified,
        full_name
    )
SELECT
    generate_gid('\x0000000000000000' :: bytea, 11),
    NOW(),
    NOW(),
    email,
    'F',
    name
FROM
    tas ON CONFLICT DO NOTHING;
