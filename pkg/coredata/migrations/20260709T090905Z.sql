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

-- Extract the SSL/ACME certificate lifecycle out of custom_domains into a
-- generic certificates table keyed on a globally-unique hostname. A custom
-- domain now references its certificate through certificate_id.

CREATE TABLE certificates (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    hostname CITEXT NOT NULL UNIQUE,
    ssl_certificate BYTEA,
    encrypted_ssl_private_key BYTEA,
    ssl_certificate_chain TEXT,
    status custom_domain_ssl_status NOT NULL,
    ssl_expires_at TIMESTAMP WITH TIME ZONE,
    ssl_retry_count INTEGER NOT NULL DEFAULT 0,
    ssl_last_attempt_at TIMESTAMP WITH TIME ZONE,
    http_challenge_token TEXT,
    http_challenge_key_auth TEXT,
    http_challenge_url TEXT,
    http_order_url TEXT,
    provisioning_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_certificates_ssl_expires ON certificates(ssl_expires_at)
    WHERE status = 'ACTIVE';
CREATE INDEX idx_certificates_http_challenge_token ON certificates(http_challenge_token)
    WHERE http_challenge_token IS NOT NULL;

-- Backfill one certificate per existing custom domain. A fresh GID is minted
-- for each row: the source domain's tenant (first 8 bytes of its GID) is kept,
-- the entity type is set to 104 (CertificateEntityType), and a millisecond
-- timestamp plus random suffix guarantee uniqueness. The certificate hostname
-- equals the custom domain name, which lets us link the two afterwards.
INSERT INTO certificates (
    id,
    tenant_id,
    hostname,
    ssl_certificate,
    encrypted_ssl_private_key,
    ssl_certificate_chain,
    status,
    ssl_expires_at,
    ssl_retry_count,
    ssl_last_attempt_at,
    http_challenge_token,
    http_challenge_key_auth,
    http_challenge_url,
    http_order_url,
    provisioning_error,
    created_at,
    updated_at
)
SELECT
    translate(
        encode(
            substring(decode(translate(cd.id, '-_', '+/'), 'base64') FROM 1 FOR 8)
            || int2send(104::smallint)
            || int8send((floor(extract(epoch FROM clock_timestamp()) * 1000))::bigint)
            || substring(decode(md5(random()::text || cd.id), 'hex') FROM 1 FOR 6),
            'base64'
        ),
        '+/',
        '-_'
    ),
    cd.tenant_id,
    cd.domain,
    cd.ssl_certificate,
    cd.encrypted_ssl_private_key,
    cd.ssl_certificate_chain,
    COALESCE(cd.ssl_status, 'PENDING'),
    cd.ssl_expires_at,
    COALESCE(cd.ssl_retry_count, 0),
    cd.ssl_last_attempt_at,
    cd.http_challenge_token,
    cd.http_challenge_key_auth,
    cd.http_challenge_url,
    cd.http_order_url,
    cd.provisioning_error,
    cd.created_at,
    cd.updated_at
FROM custom_domains cd;

ALTER TABLE custom_domains ADD COLUMN certificate_id TEXT REFERENCES certificates(id) ON DELETE SET NULL;

UPDATE custom_domains cd
SET certificate_id = c.id
FROM certificates c
WHERE c.hostname = cd.domain;

-- Repoint the certificate cache from the custom domain to the certificate.
ALTER TABLE cached_certificates ADD COLUMN certificate_id TEXT REFERENCES certificates(id) ON DELETE CASCADE;

UPDATE cached_certificates cc
SET certificate_id = cd.certificate_id
FROM custom_domains cd
WHERE cd.id = cc.custom_domain_id;

ALTER TABLE cached_certificates DROP COLUMN custom_domain_id;

-- Drop the certificate lifecycle columns now living on certificates.
DROP INDEX IF EXISTS idx_custom_domains_ssl_expires;
DROP INDEX IF EXISTS idx_custom_domains_http_challenge_token;

ALTER TABLE custom_domains
    DROP COLUMN ssl_certificate,
    DROP COLUMN encrypted_ssl_private_key,
    DROP COLUMN ssl_certificate_chain,
    DROP COLUMN ssl_status,
    DROP COLUMN ssl_expires_at,
    DROP COLUMN ssl_retry_count,
    DROP COLUMN ssl_last_attempt_at,
    DROP COLUMN http_challenge_token,
    DROP COLUMN http_challenge_key_auth,
    DROP COLUMN http_challenge_url,
    DROP COLUMN http_order_url,
    DROP COLUMN provisioning_error;
