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

ALTER TABLE processing_activities ADD COLUMN last_review_date DATE;
ALTER TABLE processing_activities ADD COLUMN next_review_date DATE;

CREATE TYPE processing_activity_role AS ENUM ('CONTROLLER', 'PROCESSOR');
ALTER TABLE processing_activities ADD COLUMN role processing_activity_role NOT NULL DEFAULT 'PROCESSOR';
ALTER TABLE processing_activities ALTER COLUMN role DROP DEFAULT;

ALTER TABLE processing_activities ADD COLUMN data_protection_officer_id TEXT REFERENCES peoples(id) ON DELETE RESTRICT;

CREATE TYPE processing_activity_dpia_residual_risk AS ENUM ('LOW', 'MEDIUM', 'HIGH');

ALTER TABLE processing_activities RENAME COLUMN data_protection_impact_assessment TO data_protection_impact_assessment_needed;
ALTER TABLE processing_activities RENAME COLUMN transfer_impact_assessment TO transfer_impact_assessment_needed;

CREATE TABLE processing_activity_data_protection_impact_assessments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    snapshot_id TEXT REFERENCES snapshots(id) ON UPDATE CASCADE ON DELETE CASCADE,
    source_id TEXT,
    organization_id TEXT NOT NULL,
    processing_activity_id TEXT NOT NULL,
    description TEXT,
    necessity_and_proportionality TEXT,
    potential_risk TEXT,
    mitigations TEXT,
    residual_risk processing_activity_dpia_residual_risk,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT processing_activity_dpia_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT processing_activity_dpia_processing_activity_id_fkey
        FOREIGN KEY (processing_activity_id)
        REFERENCES processing_activities(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT processing_activity_dpias_source_id_snapshot_id_key
        UNIQUE (source_id, snapshot_id)
);

CREATE TABLE processing_activity_transfer_impact_assessments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    snapshot_id TEXT REFERENCES snapshots(id) ON UPDATE CASCADE ON DELETE CASCADE,
    source_id TEXT,
    organization_id TEXT NOT NULL,
    processing_activity_id TEXT NOT NULL,
    data_subjects TEXT,
    legal_mechanism TEXT,
    transfer TEXT,
    local_law_risk TEXT,
    supplementary_measures TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT processing_activity_tia_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT processing_activity_tia_processing_activity_id_fkey
        FOREIGN KEY (processing_activity_id)
        REFERENCES processing_activities(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT processing_activity_tias_source_id_snapshot_id_key
        UNIQUE (source_id, snapshot_id)
);
