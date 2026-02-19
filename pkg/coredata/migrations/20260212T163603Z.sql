ALTER TABLE
    iam_membership_profiles
ADD
    COLUMN state membership_state NOT NULL DEFAULT 'ACTIVE',
ADD
    COLUMN source TEXT NOT NULL DEFAULT 'MANUAL';

UPDATE
    iam_membership_profiles p
SET
    state = m.state,
    source = m.source
FROM
    iam_memberships m
WHERE
    m.id = p.membership_id;

ALTER TABLE
    iam_membership_profiles DROP COLUMN membership_id;

ALTER TABLE
    iam_memberships
ALTER COLUMN
    source DROP NOT NULL;

ALTER TABLE
    iam_scim_events
ADD
    COLUMN user_name CITEXT NOT NULL DEFAULT '';

WITH emails AS (
    SELECT
        i.email_address,
        m.id
    FROM
        iam_memberships m
        INNER JOIN identities i ON i.id = m.identity_id
)
UPDATE
    iam_scim_events se
SET
    user_name = e.email_address
FROM
    emails e
WHERE
    e.id = se.membership_id;

ALTER TABLE
    iam_scim_events
ALTER COLUMN
    user_name DROP DEFAULT;

-- Convert invitations to identities / profiles / memberships
-- Create missing identities (one row per email to avoid "cannot affect row a second time")
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
    i.email,
    FALSE,
    i.full_name
FROM
    (
        SELECT
            DISTINCT ON (email) email,
            full_name
        FROM
            iam_invitations
        WHERE
            accepted_at IS NULL
        ORDER BY
            email,
            created_at DESC
    ) i ON CONFLICT (email_address) DO
UPDATE
SET
    full_name = COALESCE(
        NULLIF(identities.full_name, ''),
        EXCLUDED.full_name
    );

-- Create missing profiles
WITH invitation_identities AS (
    SELECT
        i.id AS identity_id,
        inv.tenant_id AS tenant_id,
        inv.organization_id AS organization_id,
        inv.full_name AS full_name
    FROM
        iam_invitations inv
        INNER JOIN identities i ON i.email_address = inv.email
    WHERE
        inv.accepted_at IS NULL
)
INSERT INTO
    iam_membership_profiles (
        id,
        tenant_id,
        identity_id,
        organization_id,
        full_name,
        kind,
        additional_email_addresses,
        source,
        state,
        created_at,
        updated_at
    )
SELECT
    generate_gid(decode_base64_unpadded(ii.tenant_id), 51),
    ii.tenant_id,
    ii.identity_id,
    ii.organization_id,
    ii.full_name,
    'EMPLOYEE',
    '{}' :: CITEXT [],
    'MANUAL',
    'INACTIVE',
    NOW(),
    NOW()
FROM
    invitation_identities ii ON CONFLICT DO NOTHING;

-- Create missing memberships
WITH invitation_identities AS (
    SELECT
        i.id AS identity_id,
        inv.tenant_id AS tenant_id,
        inv.organization_id AS organization_id,
        inv.role AS role
    FROM
        iam_invitations inv
        INNER JOIN identities i ON i.email_address = inv.email
    WHERE
        inv.accepted_at IS NULL
)
INSERT INTO
    iam_memberships (
        id,
        tenant_id,
        identity_id,
        organization_id,
        role,
        created_at,
        updated_at
    )
SELECT
    generate_gid(decode_base64_unpadded(ii.tenant_id), 39),
    ii.tenant_id,
    ii.identity_id,
    ii.organization_id,
    ii.role,
    NOW(),
    NOW()
FROM
    invitation_identities ii ON CONFLICT DO NOTHING;

ALTER TABLE
    iam_invitations
ADD
    COLUMN user_id TEXT REFERENCES iam_membership_profiles(id) ON DELETE CASCADE;

WITH profile_identities AS (
    SELECT
        p.id AS profile_id,
        i.email_address,
        p.organization_id
    FROM
        iam_membership_profiles p
        INNER JOIN identities i ON i.id = p.identity_id
)
UPDATE
    iam_invitations i
SET
    user_id = pi.profile_id
FROM
    profile_identities pi
WHERE
    pi.organization_id = i.organization_id
    AND pi.email_address = i.email;

-- Delete orphan invitations: e.g. expired invitations that were never accepted
WITH orphan_invitations AS (
    SELECT
        inv.id,
        inv.organization_id
    FROM
        iam_invitations inv
        LEFT JOIN identities i ON i.email_address = inv.email
        LEFT JOIN iam_membership_profiles p ON p.identity_id = i.id
        AND p.organization_id = inv.organization_id
    WHERE
        i.id IS NULL
        OR p.id IS NULL
)
DELETE FROM
    iam_invitations
WHERE
    id IN (
        SELECT
            id
        FROM
            orphan_invitations
    );

ALTER TABLE
    iam_invitations
ALTER COLUMN
    user_id
SET
    NOT NULL,
ALTER COLUMN
    email DROP NOT NULL,
ALTER COLUMN
    role TYPE TEXT USING role :: text,
ALTER COLUMN
    role DROP NOT NULL,
ALTER COLUMN
    full_name DROP NOT NULL;
