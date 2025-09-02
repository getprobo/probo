ALTER TABLE risks ADD COLUMN snapshot_id TEXT;
ALTER TABLE risks ADD COLUMN source_id TEXT;

ALTER TABLE risks ADD CONSTRAINT risks_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE risks ADD CONSTRAINT risks_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE risks_documents ADD COLUMN snapshot_id TEXT;

ALTER TABLE risks_documents ADD CONSTRAINT risks_documents_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE risks_measures ADD COLUMN snapshot_id TEXT;

ALTER TABLE risks_measures ADD CONSTRAINT risks_measures_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE measures ADD COLUMN snapshot_id TEXT;
ALTER TABLE measures ADD COLUMN source_id TEXT;

ALTER TABLE measures ADD CONSTRAINT measures_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE measures ADD CONSTRAINT measures_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE measures DROP CONSTRAINT mitigations_org_ref_unique;
ALTER TABLE measures ADD CONSTRAINT mitigations_org_ref_unique
    UNIQUE (organization_id, reference_id, source_id);

ALTER TABLE evidences ADD COLUMN snapshot_id TEXT;
ALTER TABLE evidences ADD COLUMN source_id TEXT;

ALTER TABLE evidences ADD CONSTRAINT evidences_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE evidences ADD CONSTRAINT evidences_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);
