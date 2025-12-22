ALTER TABLE identities ADD COLUMN full_name TEXT NOT NULL DEFAULT '';

UPDATE identities i
SET full_name = COALESCE(
    (SELECT p.full_name FROM iam_identity_profiles p 
     WHERE p.identity_id = i.id AND p.membership_id IS NULL),
    ''
);

DELETE FROM iam_identity_profiles WHERE membership_id IS NULL;
DROP INDEX IF EXISTS idx_iam_identity_profiles_default;

ALTER TABLE iam_identity_profiles DROP COLUMN identity_id;

ALTER TABLE iam_identity_profiles RENAME TO iam_membership_profiles;