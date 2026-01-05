CREATE TYPE membership_state AS ENUM ('ACTIVE', 'INACTIVE');

ALTER TABLE iam_memberships ADD COLUMN state membership_state NOT NULL DEFAULT 'ACTIVE';
CREATE INDEX idx_iam_memberships_state ON iam_memberships(state);

