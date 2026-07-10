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

-- Backfill managed default domains for compliance pages created before
-- default-domain provisioning existed. Hostnames use the managed base domain
-- suffix configured on the database (probo.trust_center_base_domain), falling
-- back to probopage.com for production installs.

WITH pending_pages AS (
    SELECT
        tc.id AS trust_center_id,
        tc.tenant_id,
        tc.organization_id,
        tc.slug,
        tc.created_at,
        (
            tc.slug || '.' || COALESCE(
                NULLIF(current_setting('probo.trust_center_base_domain', true), ''),
                'probopage.com'
            )
        )::citext AS hostname
    FROM trust_centers tc
    WHERE tc.default_domain_id IS NULL
),
minted_certificates AS (
    INSERT INTO certificates (
        id,
        tenant_id,
        hostname,
        status,
        ssl_retry_count,
        created_at,
        updated_at
    )
    SELECT
        translate(
            encode(
                substring(decode(translate(pp.trust_center_id, '-_', '+/'), 'base64') FROM 1 FOR 8)
                || int2send(104::smallint)
                || int8send((floor(extract(epoch FROM clock_timestamp()) * 1000))::bigint)
                || substring(decode(md5(random()::text || pp.trust_center_id), 'hex') FROM 1 FOR 6),
                'base64'
            ),
            '+/',
            '-_'
        ),
        pp.tenant_id,
        pp.hostname,
        'PENDING'::custom_domain_ssl_status,
        0,
        pp.created_at,
        clock_timestamp()
    FROM pending_pages pp
    RETURNING id, hostname, tenant_id
),
minted_domains AS (
    INSERT INTO custom_domains (
        id,
        tenant_id,
        organization_id,
        domain,
        managed,
        certificate_id,
        created_at,
        updated_at
    )
    SELECT
        translate(
            encode(
                substring(decode(translate(mc.id, '-_', '+/'), 'base64') FROM 1 FOR 8)
                || int2send(37::smallint)
                || int8send((floor(extract(epoch FROM clock_timestamp()) * 1000))::bigint)
                || substring(decode(md5(random()::text || mc.id), 'hex') FROM 1 FOR 6),
                'base64'
            ),
            '+/',
            '-_'
        ),
        mc.tenant_id,
        pp.organization_id,
        pp.hostname,
        true,
        mc.id,
        pp.created_at,
        clock_timestamp()
    FROM minted_certificates mc
    JOIN pending_pages pp ON pp.hostname = mc.hostname
    RETURNING id, domain
)
UPDATE trust_centers tc
SET
    default_domain_id = md.id,
    updated_at = clock_timestamp()
FROM minted_domains md
JOIN pending_pages pp ON pp.hostname = md.domain
WHERE tc.id = pp.trust_center_id;
