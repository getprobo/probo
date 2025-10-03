ALTER TABLE assets DROP CONSTRAINT assets_owner_id_fkey;
ALTER TABLE assets ADD CONSTRAINT assets_owner_id_fkey
    FOREIGN KEY (owner_id)
    REFERENCES peoples(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
