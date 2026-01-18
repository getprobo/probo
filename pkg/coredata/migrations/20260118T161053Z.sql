ALTER TABLE iam_invitations
    ALTER COLUMN expires_at TYPE TIMESTAMP WITH TIME ZONE
    USING expires_at AT TIME ZONE 'UTC';

ALTER TABLE iam_invitations
    ALTER COLUMN accepted_at TYPE TIMESTAMP WITH TIME ZONE
    USING accepted_at AT TIME ZONE 'UTC';

ALTER TABLE iam_invitations
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE
    USING created_at AT TIME ZONE 'UTC';

