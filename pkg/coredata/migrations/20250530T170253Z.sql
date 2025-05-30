CREATE TYPE document_type AS ENUM ('OTHER', 'ISMS', 'POLICY');

ALTER TABLE documents ADD COLUMN document_type document_type NOT NULL DEFAULT 'POLICY';
