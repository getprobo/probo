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

CREATE TYPE business_function_classifications AS ENUM (
    'CRITICAL',
    'IMPORTANT',
    'SECONDARY',
    'STANDARD'
);

CREATE TABLE business_functions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    reference_id TEXT NOT NULL,
    name TEXT NOT NULL,
    classification business_function_classifications NOT NULL,
    mtd_minutes INTEGER NOT NULL,
    rto_minutes INTEGER NOT NULL,
    rpo_minutes INTEGER NOT NULL,
    impact_tolerance TEXT,
    notes TEXT,
    owner_id TEXT REFERENCES iam_membership_profiles(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (organization_id, reference_id)
);

CREATE TABLE business_function_assets (
    tenant_id TEXT NOT NULL,
    business_function_id TEXT NOT NULL REFERENCES business_functions(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES assets(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (business_function_id, asset_id)
);

CREATE TABLE business_function_third_parties (
    tenant_id TEXT NOT NULL,
    business_function_id TEXT NOT NULL REFERENCES business_functions(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    third_party_id TEXT NOT NULL REFERENCES third_parties(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (business_function_id, third_party_id)
);

ALTER TABLE generated_documents
    ADD COLUMN business_functions_document_id TEXT REFERENCES documents(id) ON DELETE SET NULL;
