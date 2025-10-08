ALTER TABLE organizations
    ADD COLUMN horizontal_logo_file_id TEXT,
    ADD CONSTRAINT organizations_horizontal_logo_file_id_fkey
        FOREIGN KEY (horizontal_logo_file_id)
        REFERENCES files(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT;
