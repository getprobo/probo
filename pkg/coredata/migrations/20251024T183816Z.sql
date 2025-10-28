ALTER TABLE trust_center_accesses
ADD COLUMN last_token_expires_at TIMESTAMP WITH TIME ZONE;

UPDATE trust_center_accesses
SET last_token_expires_at = created_at + INTERVAL '7 days'
WHERE active = true AND last_token_expires_at IS NULL;

ALTER TABLE trust_center_document_accesses ADD COLUMN requested BOOLEAN NOT NULL DEFAULT false;
UPDATE trust_center_document_accesses SET requested = true WHERE active = false;
ALTER TABLE trust_center_document_accesses ALTER COLUMN requested DROP DEFAULT;
