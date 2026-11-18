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

CREATE TABLE trust_center_documents (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id) ON UPDATE CASCADE ON DELETE CASCADE,
    document_id TEXT NOT NULL REFERENCES documents(id) ON UPDATE CASCADE ON DELETE CASCADE,
    visibility trust_center_visibility NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (trust_center_id, document_id),
    CHECK (visibility IN ('PRIVATE', 'PUBLIC'))
);

CREATE TABLE trust_center_audits (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id) ON UPDATE CASCADE ON DELETE CASCADE,
    audit_id TEXT NOT NULL REFERENCES audits(id) ON UPDATE CASCADE ON DELETE CASCADE,
    visibility trust_center_visibility NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (trust_center_id, audit_id),
    CHECK (visibility IN ('PRIVATE', 'PUBLIC'))
);

CREATE TABLE trust_center_third_parties (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id) ON UPDATE CASCADE ON DELETE CASCADE,
    third_party_id TEXT NOT NULL REFERENCES third_parties(id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (trust_center_id, third_party_id)
);

INSERT INTO trust_center_documents (
    id,
    tenant_id,
    organization_id,
    trust_center_id,
    document_id,
    visibility,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(d.tenant_id), 113),
    d.tenant_id,
    d.organization_id,
    tc.id,
    d.id,
    d.trust_center_visibility,
    NOW(),
    NOW()
FROM documents d
JOIN trust_centers tc ON tc.organization_id = d.organization_id
WHERE d.trust_center_visibility IN ('PRIVATE', 'PUBLIC')
    AND d.deleted_at IS NULL;

INSERT INTO trust_center_audits (
    id,
    tenant_id,
    organization_id,
    trust_center_id,
    audit_id,
    visibility,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(a.tenant_id), 111),
    a.tenant_id,
    a.organization_id,
    tc.id,
    a.id,
    a.trust_center_visibility,
    NOW(),
    NOW()
FROM audits a
JOIN trust_centers tc ON tc.organization_id = a.organization_id
WHERE a.trust_center_visibility IN ('PRIVATE', 'PUBLIC');

INSERT INTO trust_center_third_parties (
    id,
    tenant_id,
    organization_id,
    trust_center_id,
    third_party_id,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(tp.tenant_id), 112),
    tp.tenant_id,
    tp.organization_id,
    tc.id,
    tp.id,
    NOW(),
    NOW()
FROM third_parties tp
JOIN trust_centers tc ON tc.organization_id = tp.organization_id
WHERE tp.show_on_trust_center = true;

ALTER TABLE documents DROP COLUMN trust_center_visibility;

ALTER TABLE audits DROP COLUMN trust_center_visibility;

ALTER TABLE third_parties DROP COLUMN show_on_trust_center;
