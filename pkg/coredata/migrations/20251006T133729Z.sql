ALTER TABLE
    evidences
    ADD CONSTRAINT fk_evidence_file
    FOREIGN KEY (evidence_file_id)
    REFERENCES files(id) ON DELETE SET NULL;
