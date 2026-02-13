ALTER TABLE
    iam_memberships
ADD
    COLUMN profile_id REFERENCES iam_membership_profiles(id);

UPDATE
    iam_memberships m
SET
    profile_id = p.id
FROM
    iam_membership_profiles p
WHERE
    p.membership_id = m.id;

DROP INDEX authz_memberships_user_id_organization_id_key;

DROP INDEX idx_iam_memberships_identity_tenant_org;

CREATE UNIQUE INDEX idx_iam_memberships_profile_id_idx ON iam_memberships(profile_id);

ALTER TABLE
    iam_memberships
ALTER COLUMN
    profile_id
SET
    NOT NULL,
    DROP COLUMN organization_id,
    DROP COLUMN identity_id;

ALTER TABLE
    iam_membership_profiles DROP COLUMN membership_id;
