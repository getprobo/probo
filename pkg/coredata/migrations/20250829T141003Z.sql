ALTER TYPE snapshots_type ADD VALUE 'PROCESSING_ACTIVITY_REGISTRIES';

ALTER TABLE processing_activity_registries ADD COLUMN snapshot_id TEXT;
ALTER TABLE processing_activity_registries ADD COLUMN source_id TEXT;

ALTER TABLE processing_activity_registries ADD CONSTRAINT processing_activity_registries_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE processing_activity_registries ADD CONSTRAINT processing_activity_registries_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);
