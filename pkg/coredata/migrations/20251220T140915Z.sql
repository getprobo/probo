CREATE TABLE iam_identity_profiles (
    tenant_id        TEXT NOT NULL,
    id               TEXT NOT NULL PRIMARY KEY,
    identity_id      TEXT NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    membership_id    TEXT REFERENCES iam_memberships(id) ON DELETE CASCADE,
    full_name        TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX idx_iam_identity_profiles_default ON iam_identity_profiles(identity_id) WHERE membership_id IS NULL;
CREATE UNIQUE INDEX idx_iam_identity_profiles_membership ON iam_identity_profiles(membership_id) WHERE membership_id IS NOT NULL;
