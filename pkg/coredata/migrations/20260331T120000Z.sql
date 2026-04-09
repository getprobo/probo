-- Add document_type column to document_versions, copying from parent document.
ALTER TABLE document_versions ADD COLUMN document_type document_type NOT NULL DEFAULT 'OTHER';

UPDATE document_versions dv
SET document_type = d.document_type
FROM documents d
WHERE dv.document_id = d.id;

ALTER TABLE document_versions ALTER COLUMN document_type DROP DEFAULT;

ALTER TABLE documents ALTER COLUMN document_type SET DEFAULT 'OTHER';
