-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
