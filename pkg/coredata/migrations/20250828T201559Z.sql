ALTER TABLE compliance_registries ADD COLUMN snapshot_id TEXT;
ALTER TABLE compliance_registries ADD COLUMN source_id TEXT;

ALTER TABLE compliance_registries ADD CONSTRAINT compliance_registries_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE compliance_registries ADD CONSTRAINT compliance_registries_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);
