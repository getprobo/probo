-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
