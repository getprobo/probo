ALTER TABLE risks ADD COLUMN snapshot_id TEXT;
ALTER TABLE risks ADD COLUMN source_id TEXT;

ALTER TABLE risks ADD CONSTRAINT risks_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE risks ADD CONSTRAINT risks_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);
