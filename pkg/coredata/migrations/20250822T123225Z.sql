ALTER TABLE data ADD COLUMN snapshot_id TEXT;
ALTER TABLE data ADD COLUMN original_id TEXT;

ALTER TABLE data ADD CONSTRAINT data_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE vendors ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendors ADD COLUMN original_id TEXT;

ALTER TABLE vendors ADD CONSTRAINT vendors_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE data_vendors ADD COLUMN snapshot_id TEXT;

ALTER TABLE data_vendors ADD CONSTRAINT data_vendors_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;
