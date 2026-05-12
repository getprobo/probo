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

CREATE TABLE policy_versions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    policy_id TEXT NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    content TEXT NOT NULL,
    changelog TEXT NOT NULL,
    created_by TEXT NOT NULL,
    status policy_status NOT NULL,
    published_by TEXT REFERENCES peoples(id),
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(policy_id, version_number)
);

INSERT INTO policy_versions (
    id,
    tenant_id,
    policy_id,
    version_number,
    content,
    changelog,
    created_by,
    status,
    published_by,
    published_at,
    created_at,
    updated_at
)
SELECT 
    generate_gid(decode_base64_unpadded(p.tenant_id), 16),
    p.tenant_id,
    p.id,
    1,
    p.content,
    'Initial version',
    p.owner_id,
    'DRAFT',
    NULL,
    NULL,
    p.created_at,
    p.updated_at
FROM policies p;

ALTER TABLE policies RENAME COLUMN name TO title;
ALTER TABLE policies DROP COLUMN content;
ALTER TABLE policies ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE policies ALTER COLUMN owner_id DROP NOT NULL;
ALTER TABLE policies DROP COLUMN review_date;
ALTER TABLE policies DROP COLUMN status;
ALTER TABLE policies ADD COLUMN current_published_version INTEGER;

UPDATE policies p
SET current_published_version = 1
WHERE EXISTS (
    SELECT 1 FROM policy_versions pv 
    WHERE pv.policy_id = p.id 
    AND pv.status = 'ACTIVE'
);
