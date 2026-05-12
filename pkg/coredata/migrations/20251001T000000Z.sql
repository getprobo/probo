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

DROP INDEX IF EXISTS idx_custom_domains_domain;
DROP INDEX IF EXISTS idx_custom_domains_ssl_expires;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS is_active;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS organization_id;
ALTER TABLE organizations ADD COLUMN custom_domain_id TEXT REFERENCES custom_domains(id) ON DELETE SET NULL;

CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX idx_custom_domains_ssl_expires ON custom_domains(ssl_expires_at)
    WHERE ssl_status = 'ACTIVE';
CREATE INDEX idx_organizations_custom_domain ON organizations(custom_domain_id) WHERE custom_domain_id IS NOT NULL;
