-- Access Review feature: enums and tables

-- Enum types
CREATE TYPE access_review_campaign_status AS ENUM (
    'DRAFT',
    'IN_PROGRESS',
    'PENDING_ACTIONS',
    'COMPLETED',
    'CANCELLED'
);

CREATE TYPE access_source_category AS ENUM (
    'SAAS',
    'CLOUD_INFRA',
    'SOURCE_CODE',
    'OTHER'
);

CREATE TYPE access_entry_flag AS ENUM (
    'NONE',
    'ORPHANED',
    'INACTIVE',
    'EXCESSIVE',
    'ROLE_MISMATCH',
    'NEW'
);

CREATE TYPE access_entry_decision AS ENUM (
    'PENDING',
    'APPROVED',
    'REVOKE',
    'MODIFY'
);

CREATE TYPE mfa_status AS ENUM (
    'ENABLED',
    'DISABLED',
    'UNKNOWN'
);

CREATE TYPE auth_method AS ENUM (
    'SSO',
    'PASSWORD',
    'API_KEY',
    'SERVICE_ACCOUNT',
    'UNKNOWN'
);

-- 1. access_reviews: one per organization, per-org access review configuration
CREATE TABLE access_reviews (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    identity_source_id TEXT, -- FK added after access_sources table is created
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_reviews_tenant_id ON access_reviews(tenant_id);
CREATE INDEX idx_access_reviews_organization_id ON access_reviews(organization_id);

-- 2. access_sources: configured data sources, owned by the access review
CREATE TABLE access_sources (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_id TEXT NOT NULL REFERENCES access_reviews(id) ON DELETE CASCADE,
    connector_id TEXT REFERENCES connectors(id),
    name TEXT NOT NULL,
    category access_source_category NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_sources_tenant_id ON access_sources(tenant_id);
CREATE INDEX idx_access_sources_access_review_id ON access_sources(access_review_id);
CREATE INDEX idx_access_sources_connector_id ON access_sources(connector_id);

-- Now add the identity_source_id FK constraint
ALTER TABLE access_reviews
    ADD CONSTRAINT fk_access_reviews_identity_source
    FOREIGN KEY (identity_source_id) REFERENCES access_sources(id) ON DELETE SET NULL;

-- 3. access_review_campaigns: individual review campaigns
CREATE TABLE access_review_campaigns (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_id TEXT NOT NULL REFERENCES access_reviews(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    status access_review_campaign_status NOT NULL DEFAULT 'DRAFT',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    framework_controls TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_review_campaigns_tenant_id ON access_review_campaigns(tenant_id);
CREATE INDEX idx_access_review_campaigns_access_review_id ON access_review_campaigns(access_review_id);
CREATE INDEX idx_access_review_campaigns_status ON access_review_campaigns(status);

-- 4. access_review_campaign_scope_systems: join table campaigns <-> access_sources
CREATE TABLE access_review_campaign_scope_systems (
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT NOT NULL REFERENCES access_sources(id) ON DELETE CASCADE,
    PRIMARY KEY (access_review_campaign_id, access_source_id)
);

-- 5. access_review_campaign_reviewers: join table campaigns <-> peoples
CREATE TABLE access_review_campaign_reviewers (
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    people_id TEXT NOT NULL REFERENCES peoples(id) ON DELETE CASCADE,
    PRIMARY KEY (access_review_campaign_id, people_id)
);

-- 6. access_entries: individual access records per user per system per campaign
CREATE TABLE access_entries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT NOT NULL REFERENCES access_sources(id) ON DELETE CASCADE,
    identity_id TEXT REFERENCES identities(id) ON DELETE SET NULL,
    email TEXT NOT NULL,
    full_name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT '',
    job_title TEXT NOT NULL DEFAULT '',
    is_admin BOOLEAN NOT NULL DEFAULT false,
    mfa_status mfa_status NOT NULL DEFAULT 'UNKNOWN',
    auth_method auth_method NOT NULL DEFAULT 'UNKNOWN',
    last_login TIMESTAMP WITH TIME ZONE,
    account_created_at TIMESTAMP WITH TIME ZONE,
    external_id TEXT NOT NULL DEFAULT '',
    flag access_entry_flag NOT NULL DEFAULT 'NONE',
    flag_reason TEXT,
    decision access_entry_decision NOT NULL DEFAULT 'PENDING',
    decision_note TEXT,
    decided_by TEXT REFERENCES peoples(id) ON DELETE SET NULL,
    decided_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_entries_tenant_id ON access_entries(tenant_id);
CREATE INDEX idx_access_entries_campaign_id ON access_entries(access_review_campaign_id);
CREATE INDEX idx_access_entries_source_id ON access_entries(access_source_id);
CREATE INDEX idx_access_entries_identity_id ON access_entries(identity_id);
CREATE INDEX idx_access_entries_flag ON access_entries(flag);
CREATE INDEX idx_access_entries_decision ON access_entries(decision);
