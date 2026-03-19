ALTER TYPE policy_status RENAME TO document_version_status;

CREATE TYPE document_status AS ENUM ('ACTIVE', 'ARCHIVED');

ALTER TABLE documents ADD COLUMN archived_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE documents ADD COLUMN status document_status NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE documents ALTER COLUMN status DROP DEFAULT;
