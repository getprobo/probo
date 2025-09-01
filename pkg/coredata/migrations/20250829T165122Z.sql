ALTER TABLE vendor_business_associate_agreements ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendor_business_associate_agreements ADD COLUMN source_id TEXT;

ALTER TABLE vendor_business_associate_agreements DROP CONSTRAINT vendor_business_associate_agreements_file_id_unique;
ALTER TABLE vendor_business_associate_agreements DROP CONSTRAINT vendor_business_associate_agreements_file_id_fkey;
ALTER TABLE vendor_business_associate_agreements ADD CONSTRAINT vendor_business_associate_agreements_file_id_fkey
    FOREIGN KEY (file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;

ALTER TABLE vendor_business_associate_agreements ADD CONSTRAINT vendor_business_associate_agreements_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE vendor_business_associate_agreements ADD CONSTRAINT vendor_business_associate_agreements_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE vendor_contacts ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendor_contacts ADD COLUMN source_id TEXT;

ALTER TABLE vendor_contacts ADD CONSTRAINT vendor_contacts_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE vendor_contacts ADD CONSTRAINT vendor_contacts_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE vendor_data_privacy_agreements ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendor_data_privacy_agreements ADD COLUMN source_id TEXT;

ALTER TABLE vendor_data_privacy_agreements DROP CONSTRAINT vendor_data_privacy_agreements_file_id_unique;
ALTER TABLE vendor_data_privacy_agreements DROP CONSTRAINT vendor_data_privacy_agreements_file_id_fkey;
ALTER TABLE vendor_data_privacy_agreements ADD CONSTRAINT vendor_data_privacy_agreements_file_id_fkey
    FOREIGN KEY (file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;

ALTER TABLE vendor_data_privacy_agreements ADD CONSTRAINT vendor_data_privacy_agreements_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE vendor_data_privacy_agreements ADD CONSTRAINT vendor_data_privacy_agreements_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE vendor_risk_assessments ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendor_risk_assessments ADD COLUMN source_id TEXT;

ALTER TABLE vendor_risk_assessments ADD CONSTRAINT vendor_risk_assessments_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE vendor_risk_assessments ADD CONSTRAINT vendor_risk_assessments_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE vendor_services ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendor_services ADD COLUMN source_id TEXT;

ALTER TABLE vendor_services ADD CONSTRAINT vendor_services_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE vendor_services ADD CONSTRAINT vendor_services_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE vendor_compliance_reports ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendor_compliance_reports ADD COLUMN source_id TEXT;

ALTER TABLE vendor_compliance_reports ADD CONSTRAINT vendor_compliance_reports_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;
