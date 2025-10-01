DROP INDEX IF EXISTS idx_custom_domains_domain;
DROP INDEX IF EXISTS idx_custom_domains_ssl_expires;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS is_active;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS organization_id;
ALTER TABLE organizations ADD COLUMN custom_domain_id TEXT REFERENCES custom_domains(id) ON DELETE SET NULL;

CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX idx_custom_domains_ssl_expires ON custom_domains(ssl_expires_at)
    WHERE ssl_status = 'ACTIVE';
CREATE INDEX idx_organizations_custom_domain ON organizations(custom_domain_id) WHERE custom_domain_id IS NOT NULL;
