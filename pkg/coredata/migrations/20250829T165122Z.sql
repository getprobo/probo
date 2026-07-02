-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

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
