ALTER TABLE auth_saml_configurations DROP COLUMN enabled;
ALTER TABLE auth_saml_configurations DROP COLUMN domain_verified;

CREATE UNIQUE INDEX idx_saml_config_domain_org_unique
    ON auth_saml_configurations(organization_id, email_domain);
