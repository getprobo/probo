ALTER TABLE
    evidences
    ADD CONSTRAINT fk_evidence_file
    FOREIGN KEY (evidence_file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
