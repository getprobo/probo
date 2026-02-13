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
