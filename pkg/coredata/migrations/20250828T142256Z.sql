ALTER TABLE nonconformity_registries ADD COLUMN snapshot_id TEXT;
ALTER TABLE nonconformity_registries ADD COLUMN source_id TEXT;

ALTER TABLE nonconformity_registries DROP CONSTRAINT IF EXISTS nonconformity_registries_audit_id_fkey;
ALTER TABLE nonconformity_registries ADD CONSTRAINT nonconformity_registries_audit_id_fkey
    FOREIGN KEY (audit_id) REFERENCES audits(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE nonconformity_registries ADD CONSTRAINT nonconformity_registries_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE nonconformity_registries ADD CONSTRAINT nonconformity_registries_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);
