ALTER TABLE
    frameworks
ADD
    COLUMN light_logo_file_id TEXT,
ADD
    COLUMN dark_logo_file_id TEXT;

ALTER TABLE
    frameworks
ADD
    CONSTRAINT frameworks_light_logo_file_id_fkey FOREIGN KEY (light_logo_file_id) REFERENCES files(id) ON DELETE
SET
    NULL,
ADD
    CONSTRAINT frameworks_dark_logo_file_id_fkey FOREIGN KEY (dark_logo_file_id) REFERENCES files(id) ON DELETE
SET
    NULL;