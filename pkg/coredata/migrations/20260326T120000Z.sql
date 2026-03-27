ALTER TABLE document_versions DROP CONSTRAINT document_versions_document_id_version_number_key;

ALTER TABLE document_versions RENAME COLUMN version_number TO major;
ALTER TABLE document_versions ADD COLUMN minor INTEGER NOT NULL DEFAULT 0;
ALTER TABLE document_versions ALTER COLUMN minor DROP DEFAULT;

ALTER TABLE document_versions ADD CONSTRAINT document_versions_document_id_major_minor_key UNIQUE (document_id, major, minor);

ALTER TABLE documents RENAME COLUMN current_published_version TO current_published_major;
ALTER TABLE documents ADD COLUMN current_published_minor INTEGER;
UPDATE documents SET current_published_minor = 0 WHERE current_published_major IS NOT NULL;
