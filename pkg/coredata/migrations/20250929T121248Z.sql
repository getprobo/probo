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

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE custom_domain_ssl_status AS ENUM (
    'PENDING',
    'PROVISIONING',
    'ACTIVE',
    'RENEWING',
    'EXPIRED',
    'FAILED'
);

CREATE TABLE custom_domains (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    domain CITEXT NOT NULL UNIQUE,
    ssl_certificate BYTEA,
    encrypted_ssl_private_key BYTEA,
    ssl_certificate_chain TEXT,
    ssl_status custom_domain_ssl_status,
    ssl_expires_at TIMESTAMP WITH TIME ZONE,
    http_challenge_token TEXT,
    http_challenge_key_auth TEXT,
    http_challenge_url TEXT,
    http_order_url TEXT,
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNLOGGED TABLE cached_certificates (
    domain CITEXT PRIMARY KEY,
    certificate_pem TEXT NOT NULL,
    private_key_pem TEXT NOT NULL,
    certificate_chain TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    cached_at TIMESTAMP WITH TIME ZONE NOT NULL,
    custom_domain_id TEXT REFERENCES custom_domains(id) ON DELETE CASCADE
);

CREATE INDEX idx_custom_domains_domain ON custom_domains(domain) WHERE is_active = true;
CREATE INDEX idx_custom_domains_org ON custom_domains(organization_id);
CREATE INDEX idx_custom_domains_ssl_expires ON custom_domains(ssl_expires_at)
    WHERE ssl_status = 'ACTIVE' AND is_active = true;
CREATE INDEX idx_custom_domains_http_challenge_token ON custom_domains(http_challenge_token)
    WHERE http_challenge_token IS NOT NULL;
CREATE INDEX idx_certificate_cache_expires ON cached_certificates(expires_at);
