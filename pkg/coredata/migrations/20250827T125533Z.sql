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

CREATE TYPE processing_activity_registries_special_or_criminal_data AS ENUM (
    'YES',
    'NO',
    'POSSIBLE'
);

CREATE TYPE processing_activity_registries_lawful_basis AS ENUM (
    'LEGITIMATE_INTEREST',
    'CONSENT',
    'CONTRACTUAL_NECESSITY',
    'LEGAL_OBLIGATION',
    'VITAL_INTERESTS',
    'PUBLIC_TASK'
);

CREATE TYPE processing_activity_registries_transfer_safeguards AS ENUM (
    'STANDARD_CONTRACTUAL_CLAUSES',
    'BINDING_CORPORATE_RULES',
    'ADEQUACY_DECISION',
    'DEROGATIONS',
    'CODES_OF_CONDUCT',
    'CERTIFICATION_MECHANISMS'
);

CREATE TYPE processing_activity_registries_data_protection_impact_assessment AS ENUM (
    'NEEDED',
    'NOT_NEEDED'
);

CREATE TYPE processing_activity_registries_transfer_impact_assessment AS ENUM (
    'NEEDED',
    'NOT_NEEDED'
);

CREATE TABLE processing_activity_registries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    name TEXT NOT NULL,
    purpose TEXT,
    data_subject_category TEXT,
    personal_data_category TEXT,
    special_or_criminal_data processing_activity_registries_special_or_criminal_data NOT NULL,
    consent_evidence_link TEXT,
    lawful_basis processing_activity_registries_lawful_basis NOT NULL,
    recipients TEXT,
    location TEXT,
    international_transfers BOOLEAN NOT NULL,
    transfer_safeguards processing_activity_registries_transfer_safeguards,
    retention_period TEXT,
    security_measures TEXT,
    data_protection_impact_assessment processing_activity_registries_data_protection_impact_assessment NOT NULL,
    transfer_impact_assessment processing_activity_registries_transfer_impact_assessment NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT processing_activity_registries_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);
