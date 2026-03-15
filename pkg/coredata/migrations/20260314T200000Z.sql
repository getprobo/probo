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

CREATE TYPE access_review_remediation_task_status AS ENUM (
    'OPEN',
    'ACKNOWLEDGED',
    'COMPLETED',
    'CANCELLED'
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

-- 1. access_reviews: one per organization
CREATE TABLE access_reviews (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    identity_source_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_reviews_tenant_id ON access_reviews(tenant_id);
CREATE INDEX idx_access_reviews_organization_id ON access_reviews(organization_id);

-- 2. access_sources: configured data sources
CREATE TABLE access_sources (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_id TEXT NOT NULL REFERENCES access_reviews(id) ON DELETE CASCADE,
    connector_id TEXT REFERENCES connectors(id),
    name TEXT NOT NULL,
    category access_source_category NOT NULL,
    csv_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_sources_tenant_id ON access_sources(tenant_id);
CREATE INDEX idx_access_sources_access_review_id ON access_sources(access_review_id);
CREATE INDEX idx_access_sources_connector_id ON access_sources(connector_id);

-- FK for identity_source_id now that access_sources exists
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

-- 5. access_entries: individual access records per user per system per campaign
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

-- 6. access_entry_decision_events: audit trail for decision changes
CREATE TABLE access_entry_decision_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_entry_id TEXT NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    previous_decision access_entry_decision NOT NULL,
    new_decision access_entry_decision NOT NULL,
    decision_note TEXT,
    decided_by TEXT,
    decided_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_entry_decision_events_tenant_id
    ON access_entry_decision_events (tenant_id);
CREATE INDEX idx_access_entry_decision_events_campaign_id
    ON access_entry_decision_events (access_review_campaign_id);
CREATE INDEX idx_access_entry_decision_events_entry_id
    ON access_entry_decision_events (access_entry_id);

-- 7. access_review_remediation_tasks: follow-up tasks for revoked/escalated entries
CREATE TABLE access_review_remediation_tasks (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_entry_id TEXT NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,
    decided_by TEXT,
    status access_review_remediation_task_status NOT NULL DEFAULT 'OPEN',
    status_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_entry_id)
);

CREATE INDEX idx_access_review_remediation_tasks_tenant_id
    ON access_review_remediation_tasks (tenant_id);
CREATE INDEX idx_access_review_remediation_tasks_campaign_id
    ON access_review_remediation_tasks (access_review_campaign_id);
CREATE INDEX idx_access_review_remediation_tasks_status
    ON access_review_remediation_tasks (status);

-- 8. access_review_campaign_validation_checkpoints
CREATE TABLE access_review_campaign_validation_checkpoints (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    note TEXT,
    validated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_campaign_id)
);

CREATE INDEX idx_access_review_campaign_validation_checkpoints_tenant_id
    ON access_review_campaign_validation_checkpoints (tenant_id);

-- 9. access_review_campaign_evidence_snapshots
CREATE TABLE access_review_campaign_evidence_snapshots (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    checksum_sha256 TEXT NOT NULL,
    signature TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_campaign_id)
);

CREATE INDEX idx_access_review_campaign_evidence_snapshots_tenant_id
    ON access_review_campaign_evidence_snapshots (tenant_id);

-- 10. access_review_campaign_source_fetches: tracks per-source fetch lifecycle
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
