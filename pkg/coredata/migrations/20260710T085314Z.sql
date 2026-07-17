-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

ALTER TABLE trust_centers
    ADD COLUMN description TEXT,
    ADD COLUMN website_url TEXT,
    ADD COLUMN email CITEXT,
    ADD COLUMN headquarter_address TEXT;

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
FROM organizations o
WHERE NOT EXISTS (
        SELECT 1 FROM trust_centers tc WHERE tc.organization_id = o.id
    )
    AND (
        o.description IS NOT NULL
        OR o.website_url IS NOT NULL
        OR o.email IS NOT NULL
        OR o.headquarter_address IS NOT NULL
    );

UPDATE trust_centers tc
SET
    description = o.description,
    website_url = o.website_url,
    email = o.email,
    headquarter_address = o.headquarter_address
FROM organizations o
WHERE tc.organization_id = o.id;

ALTER TABLE organizations
    DROP COLUMN description,
    DROP COLUMN website_url,
    DROP COLUMN email,
    DROP COLUMN headquarter_address;
