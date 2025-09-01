ALTER TABLE assets ADD COLUMN snapshot_id TEXT;
ALTER TABLE assets ADD COLUMN source_id TEXT;

ALTER TABLE assets ADD CONSTRAINT assets_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE assets ADD CONSTRAINT assets_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE asset_vendors ADD COLUMN snapshot_id TEXT;

ALTER TABLE asset_vendors ADD CONSTRAINT asset_vendors_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;
