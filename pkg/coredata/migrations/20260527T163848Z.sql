-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

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
