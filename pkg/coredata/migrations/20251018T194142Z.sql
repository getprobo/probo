CREATE TYPE saml_enforcement_policy AS ENUM (
    'OFF',
    'OPTIONAL',
    'REQUIRED'
);

CREATE TABLE auth_saml_configurations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    email_domain TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    enforcement_policy saml_enforcement_policy NOT NULL,
    idp_entity_id TEXT NOT NULL,
    idp_sso_url TEXT NOT NULL,
    idp_certificate TEXT NOT NULL,
    idp_metadata_url TEXT,  
    attribute_email TEXT NOT NULL DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress',
    attribute_firstname TEXT NOT NULL DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname',
    attribute_lastname TEXT NOT NULL DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname',
    attribute_role TEXT NOT NULL DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role',
    default_role TEXT NOT NULL DEFAULT 'MEMBER',
    auto_signup_enabled BOOLEAN NOT NULL DEFAULT false,
    domain_verified BOOLEAN NOT NULL DEFAULT false,
    domain_verification_token TEXT,
    domain_verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_auth_saml_configurations_organization FOREIGN KEY (organization_id)
        REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_saml_configurations_organization_id
    ON auth_saml_configurations(organization_id);

CREATE INDEX idx_auth_saml_configurations_tenant_id
    ON auth_saml_configurations(tenant_id);

CREATE UNIQUE INDEX idx_saml_config_domain_org_unique
    ON auth_saml_configurations(organization_id, email_domain)
    WHERE enabled = true AND domain_verified = true;

CREATE INDEX idx_saml_config_email_domain
    ON auth_saml_configurations(email_domain)
    WHERE enabled = true AND domain_verified = true;

ALTER TABLE users ADD COLUMN saml_subject TEXT;

CREATE UNIQUE INDEX idx_users_saml_subject
    ON users(saml_subject)
    WHERE saml_subject IS NOT NULL;

ALTER TABLE users ALTER COLUMN hashed_password DROP NOT NULL;

CREATE TABLE auth_saml_assertions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    used_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_auth_saml_assertions_organization FOREIGN KEY (organization_id)
        REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_saml_assertions_expires_at
    ON auth_saml_assertions(expires_at);

CREATE INDEX idx_auth_saml_assertions_organization_id
    ON auth_saml_assertions(organization_id);

CREATE INDEX idx_auth_saml_assertions_tenant_id
    ON auth_saml_assertions(tenant_id);

CREATE TABLE auth_saml_requests (
    id TEXT PRIMARY KEY, 
    organization_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_auth_saml_requests_organization FOREIGN KEY (organization_id)
        REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_saml_requests_id_org ON auth_saml_requests(id, organization_id);

CREATE INDEX idx_auth_saml_requests_expires_at ON auth_saml_requests(expires_at);

CREATE TABLE auth_saml_relay_states (
    token TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    request_id TEXT NOT NULL,
    saml_config_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_auth_saml_relay_states_organization FOREIGN KEY (organization_id)
        REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_saml_relay_states_saml_config FOREIGN KEY (saml_config_id)
        REFERENCES auth_saml_configurations(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_saml_relay_states_request FOREIGN KEY (request_id)
        REFERENCES auth_saml_requests(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_saml_relay_states_token ON auth_saml_relay_states(token);

CREATE INDEX idx_auth_saml_relay_states_expires_at ON auth_saml_relay_states(expires_at);

CREATE INDEX idx_auth_saml_relay_states_tenant_id ON auth_saml_relay_states(tenant_id);

ALTER TABLE auth_saml_configurations DROP COLUMN default_role;