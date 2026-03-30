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
