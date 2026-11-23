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

-- A third party can have any number of agreements (MSA, CGV/CGU, BAA,
-- DPA, amendments, annexes). The previous tables allowed one business
-- associate agreement and one data privacy agreement. Each agreement has
-- a type and an editable title; the attached file keeps its own name.

CREATE TYPE third_party_agreement_type AS ENUM (
    'AMENDMENT',
    'ANNEX',
    'BUSINESS_ASSOCIATE',
    'DATA_PRIVACY',
    'MASTER_SERVICE',
    'OTHER',
    'TERMS'
);

CREATE TABLE third_party_agreements (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations (id) ON UPDATE CASCADE ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    third_party_id TEXT NOT NULL REFERENCES third_parties (id) ON UPDATE CASCADE ON DELETE CASCADE,
    title TEXT NOT NULL,
    type third_party_agreement_type NOT NULL,
    valid_from DATE,
    valid_until DATE,
    file_id TEXT NOT NULL REFERENCES files (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- generate_gid() is defined in 20250420T120000Z. Entity type 150 is
-- ThirdPartyAgreementEntityType.
INSERT INTO third_party_agreements (
    id,
    organization_id,
    tenant_id,
    third_party_id,
    title,
    type,
    valid_from,
    valid_until,
    file_id,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(tenant_id), 150),
    organization_id,
    tenant_id,
    third_party_id,
    'Business associate agreement',
    'BUSINESS_ASSOCIATE',
    valid_from,
    valid_until,
    file_id,
    created_at,
    updated_at
FROM
    third_party_business_associate_agreements;

INSERT INTO third_party_agreements (
    id,
    organization_id,
    tenant_id,
    third_party_id,
    title,
    type,
    valid_from,
    valid_until,
    file_id,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(tenant_id), 150),
    organization_id,
    tenant_id,
    third_party_id,
    'Data privacy agreement',
    'DATA_PRIVACY',
    valid_from,
    valid_until,
    file_id,
    created_at,
    updated_at
FROM
    third_party_data_privacy_agreements;

-- TODO: Drop third_party_business_associate_agreements and
-- third_party_data_privacy_agreements in a later migration, once the
-- rows copied above have been verified. They are kept until then so
-- the original data can be checked or restored.
