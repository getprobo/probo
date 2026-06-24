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

DROP INDEX IF EXISTS idx_custom_domains_domain;
DROP INDEX IF EXISTS idx_custom_domains_ssl_expires;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS is_active;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS organization_id;
ALTER TABLE organizations ADD COLUMN custom_domain_id TEXT REFERENCES custom_domains(id) ON DELETE SET NULL;

CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX idx_custom_domains_ssl_expires ON custom_domains(ssl_expires_at)
    WHERE ssl_status = 'ACTIVE';
CREATE INDEX idx_organizations_custom_domain ON organizations(custom_domain_id) WHERE custom_domain_id IS NOT NULL;
