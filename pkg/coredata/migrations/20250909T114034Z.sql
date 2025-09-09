ALTER TABLE trust_center_accesses ADD COLUMN has_accepted_non_disclosure_agreement BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE trust_center_accesses ALTER COLUMN has_accepted_non_disclosure_agreement DROP DEFAULT;
ALTER TABLE trust_center_accesses ADD COLUMN has_accepted_non_disclosure_agreement_metadata JSONB;

ALTER TABLE trust_centers ADD COLUMN non_disclosure_agreement_file_id TEXT;
ALTER TABLE trust_centers ADD CONSTRAINT trust_centers_non_disclosure_agreement_file_id_fkey
    FOREIGN KEY (non_disclosure_agreement_file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
