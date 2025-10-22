CREATE TYPE document_classification AS ENUM (
    'PUBLIC',
    'INTERNAL',
    'CONFIDENTIAL',
    'SECRET'
);

ALTER TABLE documents ADD COLUMN classification document_classification NOT NULL DEFAULT 'CONFIDENTIAL';
ALTER TABLE documents ALTER COLUMN classification DROP DEFAULT;

ALTER TABLE document_versions ADD COLUMN classification document_classification NOT NULL DEFAULT 'CONFIDENTIAL';
ALTER TABLE document_versions ALTER COLUMN classification DROP DEFAULT;
