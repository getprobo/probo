ALTER TABLE
    iam_memberships DROP COLUMN source,
    DROP COLUMN state;

ALTER TABLE
    iam_invitations DROP COLUMN full_name,
    DROP COLUMN email,
    DROP COLUMN role;
