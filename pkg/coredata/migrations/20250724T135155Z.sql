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

CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TABLE trust_centers (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    active BOOLEAN NOT NULL,
    slug TEXT NOT NULL CHECK (slug ~ '^[a-z0-9_-]+$'),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(slug)
);

INSERT INTO trust_centers (
    id,
    organization_id,
    tenant_id,
    active,
    slug,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(o.tenant_id), 22),
    o.id,
    o.tenant_id,
    false,
    LOWER(
        REGEXP_REPLACE(
            REGEXP_REPLACE(
                unaccent(o.name),
                '[^a-zA-Z0-9\s]', '', 'g'
            ),
            '\s+', '-', 'g'
        )
    ),
    NOW(),
    NOW()
FROM organizations o;

ALTER TABLE documents ADD COLUMN show_on_trust_center BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE audits ADD COLUMN show_on_trust_center BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE vendors ADD COLUMN show_on_trust_center BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE documents ALTER COLUMN show_on_trust_center DROP DEFAULT;

ALTER TABLE audits ALTER COLUMN show_on_trust_center DROP DEFAULT;

ALTER TABLE vendors ALTER COLUMN show_on_trust_center DROP DEFAULT;
