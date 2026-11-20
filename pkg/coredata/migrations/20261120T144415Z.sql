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

-- A third party can have any number of named agreements (MSA, CGV/CGU, BAA,
-- DPA, amendments, annexes). The previous tables allowed one business
-- associate agreement and one data privacy agreement.

CREATE TYPE third_party_agreement_type AS ENUM (
    'BUSINESS_ASSOCIATE',
    'DATA_PRIVACY',
    'TERMS',
    'MASTER_SERVICE',
    'AMENDMENT',
    'ANNEX',
    'OTHER'
);

CREATE TABLE third_party_agreements (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations (id) ON UPDATE CASCADE ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    third_party_id TEXT NOT NULL REFERENCES third_parties (id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL,
    type third_party_agreement_type NOT NULL,
    valid_from DATE,
    valid_until DATE,
    file_id TEXT NOT NULL REFERENCES files (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Mint a GID (base64url of: tenant 8B | entity type 2B | timestamp 8B | random 6B).
-- Entity type 150 is ThirdPartyAgreementEntityType.
CREATE FUNCTION pg_temp.gen_gid(tenant_text text, entity_type int) RETURNS text AS $$
    SELECT translate(
        encode(
            decode(rpad(translate(tenant_text, '-_', '+/'), 12, '='), 'base64')
            || set_byte(set_byte('\x0000'::bytea, 0, (entity_type >> 8) & 255), 1, entity_type & 255)
            || int8send((extract(epoch FROM clock_timestamp()) * 1000)::bigint)
            || substring(uuid_send(gen_random_uuid()) FROM 1 FOR 6),
            'base64'),
        '+/', '-_')
$$ LANGUAGE sql VOLATILE;

INSERT INTO third_party_agreements (
    id,
    organization_id,
    tenant_id,
    third_party_id,
    name,
    type,
    valid_from,
    valid_until,
    file_id,
    created_at,
    updated_at
)
SELECT
    pg_temp.gen_gid(tenant_id, 150),
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
    name,
    type,
    valid_from,
    valid_until,
    file_id,
    created_at,
    updated_at
)
SELECT
    pg_temp.gen_gid(tenant_id, 150),
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

DROP TABLE third_party_business_associate_agreements;
DROP TABLE third_party_data_privacy_agreements;
