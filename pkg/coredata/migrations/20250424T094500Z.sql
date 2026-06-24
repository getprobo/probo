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
