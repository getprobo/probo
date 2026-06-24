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

INSERT INTO peoples (
    id,
    organization_id,
    full_name,
    primary_email_address,
    additional_email_addresses,
    user_id,
    kind,
    created_at,
    updated_at
)
SELECT 
    generate_gid(decode_base64_unpadded(o.tenant_id), 8),
    o.id,
    u.fullname,
    u.email_address,
    ARRAY[]::TEXT[],
    u.id,
    'CONTRACTOR',
    NOW(),
    NOW()
FROM users u
JOIN users_organizations uo ON u.id = uo.user_id
JOIN organizations o ON uo.organization_id = o.id
LEFT JOIN peoples p ON u.id = p.user_id
WHERE p.id IS NULL;

UPDATE peoples p
SET user_id = u.id
FROM users u
WHERE p.user_id IS NULL
AND p.primary_email_address = u.email_address;
