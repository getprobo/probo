-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

-- Drop the snapshot schema. The previous migration deleted all snapshot-scoped
-- data and the application code no longer reads or writes these columns.

-- Partial / expression indexes that reference snapshot_id must be dropped before
-- the column drops succeed.
DROP INDEX IF EXISTS processing_activity_dpias_processing_activity_id_snapshot_id_uniq;
DROP INDEX IF EXISTS processing_activity_tias_processing_activity_id_snapshot_id_uniq;
DROP INDEX IF EXISTS states_of_applicability_source_id_snapshot_id_uniq;
DROP INDEX IF EXISTS states_of_applicability_name_organization_id_uniq;
DROP INDEX IF EXISTS idx_findings_organization_snapshot;
DROP INDEX IF EXISTS idx_findings_org_reference_id;

-- Drop snapshot_id and source_id columns. CASCADE removes the FK to snapshots(id)
-- and any UNIQUE (source_id, snapshot_id) constraints.
ALTER TABLE applicability_statements DROP COLUMN snapshot_id CASCADE;
ALTER TABLE asset_third_parties      DROP COLUMN snapshot_id CASCADE;
ALTER TABLE assets                   DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE data                     DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE data_third_parties       DROP COLUMN snapshot_id CASCADE;
ALTER TABLE findings                 DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE obligations              DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE processing_activities    DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE processing_activity_data_protection_impact_assessments
    DROP COLUMN snapshot_id CASCADE,
    DROP COLUMN source_id   CASCADE;
ALTER TABLE processing_activity_transfer_impact_assessments
    DROP COLUMN snapshot_id CASCADE,
    DROP COLUMN source_id   CASCADE;
ALTER TABLE processing_activity_third_parties DROP COLUMN snapshot_id;
ALTER TABLE risks                    DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE statements_of_applicability
    DROP COLUMN snapshot_id CASCADE,
    DROP COLUMN source_id   CASCADE;
ALTER TABLE third_parties            DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE third_party_business_associate_agreements
    DROP COLUMN snapshot_id CASCADE,
    DROP COLUMN source_id   CASCADE;
ALTER TABLE third_party_compliance_reports
    DROP COLUMN snapshot_id CASCADE,
    DROP COLUMN source_id   CASCADE;
ALTER TABLE third_party_contacts     DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;
ALTER TABLE third_party_data_privacy_agreements
    DROP COLUMN snapshot_id CASCADE,
    DROP COLUMN source_id   CASCADE;
ALTER TABLE third_party_risk_assessments
    DROP COLUMN snapshot_id CASCADE,
    DROP COLUMN source_id   CASCADE;
ALTER TABLE third_party_services     DROP COLUMN snapshot_id CASCADE,
                                     DROP COLUMN source_id   CASCADE;

-- Recreate the unique constraints that previously gated on snapshot_id IS NULL.
CREATE UNIQUE INDEX statements_of_applicability_name_organization_id_uniq
    ON statements_of_applicability (name, organization_id);

CREATE UNIQUE INDEX idx_findings_org_reference_id
    ON findings (organization_id, reference_id);

CREATE UNIQUE INDEX processing_activity_dpias_processing_activity_id_uniq
    ON processing_activity_data_protection_impact_assessments (processing_activity_id);

CREATE UNIQUE INDEX processing_activity_tias_processing_activity_id_uniq
    ON processing_activity_transfer_impact_assessments (processing_activity_id);

DROP TABLE controls_snapshots;
DROP TABLE snapshots;
DROP TYPE snapshots_type;
