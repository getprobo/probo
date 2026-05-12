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
