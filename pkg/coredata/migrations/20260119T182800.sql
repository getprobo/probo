WITH old_to_new AS (
    SELECT
        tenant_id,
        id AS old_id,
        generate_gid(decode_base64_unpadded(tenant_id), 39) as new_id,
        identity_id,
        organization_id,
        role,
        created_at,
        updated_at,
        source,
        state
    FROM
        iam_memberships m
    WHERE
        extract_entity_type(parse_gid(m.id)) = 38
),
inserted_memberships AS (
    INSERT INTO
        iam_memberships (
            tenant_id,
            id,
            identity_id,
            organization_id,
            role,
            created_at,
            updated_at,
            source,
            state
        )
    SELECT
        m.tenant_id,
        m.new_id as id,
        m.identity_id,
        m.organization_id,
        m.role,
        m.created_at,
        m.updated_at,
        m.source,
        m.state
    FROM
        old_to_new m RETURNING id
),
updated_profiles AS (
    UPDATE
        iam_membership_profiles mp
    SET
        membership_id = m.new_id
    FROM
        old_to_new m
    WHERE
        mp.membership_id = m.old_id RETURNING id
),
deleted_memberships AS (
    DELETE FROM
        iam_memberships
    WHERE
        id IN (
            SELECT
                old_id
            FROM
                old_to_new
        ) RETURNING id
)
DELETE FROM
    iam_sessions
WHERE
    membership_id IN (
        SELECT
            old_id
        FROM
            old_to_new
    );
