-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

CREATE TABLE cmp_domains (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    domain CITEXT NOT NULL,
    source_promotion_requested_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT cmp_domains_organization_id_domain_key UNIQUE (organization_id, domain)
);

ALTER TABLE cookie_banners
    ADD COLUMN cmp_domain_id TEXT REFERENCES cmp_domains(id) ON DELETE SET NULL;

CREATE TABLE cmp_domain_surfaces (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cmp_domain_id TEXT NOT NULL REFERENCES cmp_domains(id) ON DELETE CASCADE,
    origin TEXT NOT NULL,
    status TEXT NOT NULL,
    discovered_by_cookie_banner_id TEXT REFERENCES cookie_banners(id) ON DELETE SET NULL,
    last_cookie_banner_scan_id TEXT REFERENCES cookie_banner_scans(id) ON DELETE SET NULL,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT cmp_domain_surfaces_cmp_domain_id_origin_key UNIQUE (cmp_domain_id, origin),
    CONSTRAINT cmp_domain_surfaces_status_check CHECK (status IN ('UNCONFIRMED', 'IGNORED', 'VENDOR_RUN'))
);

ALTER TABLE tracker_patterns
    ADD COLUMN source_cookie_banner_id TEXT REFERENCES cookie_banners(id) ON DELETE SET NULL;
