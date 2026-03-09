ALTER TABLE iam_membership_profiles
ADD COLUMN user_name TEXT,
ADD COLUMN external_id TEXT;

UPDATE iam_membership_profiles p
SET user_name = i.email_address
FROM identities i
WHERE i.id = p.identity_id AND p.source = 'SCIM';

CREATE UNIQUE INDEX idx_profiles_user_name_organization_id
ON iam_membership_profiles (user_name, organization_id)
WHERE user_name IS NOT NULL;

CREATE UNIQUE INDEX idx_profiles_external_id_organization_id
ON iam_membership_profiles (external_id, organization_id)
WHERE external_id IS NOT NULL;
