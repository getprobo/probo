ALTER TYPE snapshots_type ADD VALUE 'CONTINUAL_IMPROVEMENT_REGISTRIES';

ALTER TABLE continual_improvement_registries ADD COLUMN snapshot_id TEXT;
ALTER TABLE continual_improvement_registries ADD COLUMN source_id TEXT;

ALTER TABLE continual_improvement_registries ADD CONSTRAINT continual_improvement_registries_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE continual_improvement_registries ADD CONSTRAINT continual_improvement_registries_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);
