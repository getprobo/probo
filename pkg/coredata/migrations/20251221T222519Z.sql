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

INSERT INTO iam_identity_profiles (tenant_id, id, identity_id, membership_id, full_name, created_at, updated_at)
SELECT
    '',
    generate_gid('\x0000000000000000'::bytea, 51),
    i.id,
    NULL,
    COALESCE(i.fullname, ''),
    i.created_at,
    NOW()
FROM identities i
WHERE NOT EXISTS (
    SELECT 1 FROM iam_identity_profiles p 
    WHERE p.identity_id = i.id AND p.membership_id IS NULL
);

INSERT INTO iam_identity_profiles (tenant_id, id, identity_id, membership_id, full_name, created_at, updated_at)
SELECT
    m.tenant_id,
    generate_gid(decode_base64_unpadded(m.tenant_id), 51),
    m.identity_id,
    m.id,
    COALESCE(i.fullname, ''),
    m.created_at,
    NOW()
FROM iam_memberships m
JOIN identities i ON i.id = m.identity_id
WHERE NOT EXISTS (
    SELECT 1 FROM iam_identity_profiles p 
    WHERE p.membership_id = m.id
);

ALTER TABLE identities DROP COLUMN fullname;
