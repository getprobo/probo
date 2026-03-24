-- Access review system: enums, tables, and indexes

-- Enum types
CREATE TYPE access_review_campaign_status AS ENUM (
    'DRAFT',
    'IN_PROGRESS',
    'PENDING_ACTIONS',
    'COMPLETED',
    'CANCELLED',
    'FAILED'
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
    'DEFER',
    'ESCALATE'
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

CREATE TYPE access_entry_incremental_tag AS ENUM (
    'NEW',
    'REMOVED',
    'UNCHANGED'
);

CREATE TYPE access_review_campaign_source_fetch_status AS ENUM (
    'QUEUED',
    'FETCHING',
    'SUCCESS',
    'FAILED'
);

-- Connector provider and protocol extensions
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'LINEAR';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'FIGMA';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'ONE_PASSWORD';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'HUBSPOT';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'DOCUSIGN';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'NOTION';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'BREX';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'TALLY';
ALTER TYPE connector_provider ADD VALUE IF NOT EXISTS 'CLOUDFLARE';
ALTER TYPE connector_protocol ADD VALUE IF NOT EXISTS 'API_KEY';

-- 1. access_sources: configured data sources
CREATE TABLE access_sources (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    connector_id TEXT REFERENCES connectors(id),
    name TEXT NOT NULL,
    category access_source_category NOT NULL,
    csv_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_sources_tenant_id ON access_sources(tenant_id);
CREATE INDEX idx_access_sources_organization_id ON access_sources(organization_id);
CREATE INDEX idx_access_sources_connector_id ON access_sources(connector_id);

-- 2. access_review_campaigns: individual review campaigns
CREATE TABLE access_review_campaigns (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    name TEXT NOT NULL,
    status access_review_campaign_status NOT NULL DEFAULT 'DRAFT',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    framework_controls TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_review_campaigns_tenant_id ON access_review_campaigns(tenant_id);
CREATE INDEX idx_access_review_campaigns_organization_id ON access_review_campaigns(organization_id);
CREATE INDEX idx_access_review_campaigns_status ON access_review_campaigns(status);

-- 3. access_review_campaign_scope_systems: join table campaigns <-> access_sources
CREATE TABLE access_review_campaign_scope_systems (
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT NOT NULL REFERENCES access_sources(id) ON DELETE CASCADE,
    PRIMARY KEY (access_review_campaign_id, access_source_id)
);

-- 4. access_entries: individual access records per user per system per campaign
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
    account_key TEXT NOT NULL DEFAULT '',
    incremental_tag access_entry_incremental_tag NOT NULL DEFAULT 'NEW',
    flag access_entry_flag NOT NULL DEFAULT 'NONE',
    flag_reason TEXT,
    decision access_entry_decision NOT NULL DEFAULT 'PENDING',
    decision_note TEXT,
    decided_by TEXT,
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
CREATE UNIQUE INDEX idx_access_entries_campaign_source_account_key
    ON access_entries (access_review_campaign_id, access_source_id, account_key);
CREATE INDEX idx_access_entries_campaign_decision
    ON access_entries (access_review_campaign_id, decision);
CREATE INDEX idx_access_entries_campaign_source
    ON access_entries (access_review_campaign_id, access_source_id);
CREATE INDEX idx_access_entries_campaign_decided_at
    ON access_entries (access_review_campaign_id, decided_at);

-- 5. access_review_campaign_source_fetches: tracks per-source fetch lifecycle
CREATE TABLE access_review_campaign_source_fetches (
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT NOT NULL REFERENCES access_sources(id) ON DELETE CASCADE,
    status access_review_campaign_source_fetch_status NOT NULL DEFAULT 'QUEUED',
    fetched_accounts_count INTEGER NOT NULL DEFAULT 0,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (access_review_campaign_id, access_source_id)
);

CREATE INDEX idx_arc_source_fetches_tenant_id
    ON access_review_campaign_source_fetches (tenant_id);
CREATE INDEX idx_arc_source_fetches_status_updated_at
    ON access_review_campaign_source_fetches (status, updated_at);
CREATE INDEX idx_arc_source_fetches_campaign_id
    ON access_review_campaign_source_fetches (access_review_campaign_id);
