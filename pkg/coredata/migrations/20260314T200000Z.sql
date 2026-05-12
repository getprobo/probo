-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

-- Access review system: enums and tables

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

-- 3. access_review_campaign_scope_systems: join table campaigns <-> access_sources
CREATE TABLE access_review_campaign_scope_systems (
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT NOT NULL REFERENCES access_sources(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
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
    full_name TEXT NOT NULL,
    role TEXT NOT NULL,
    job_title TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL,
    mfa_status mfa_status NOT NULL,
    auth_method auth_method NOT NULL,
    last_login TIMESTAMP WITH TIME ZONE,
    account_created_at TIMESTAMP WITH TIME ZONE,
    external_id TEXT NOT NULL,
    account_key TEXT NOT NULL,
    incremental_tag access_entry_incremental_tag NOT NULL,
    flag access_entry_flag NOT NULL,
    flag_reason TEXT,
    decision access_entry_decision NOT NULL,
    decision_note TEXT,
    decided_by TEXT,
    decided_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_access_entries_campaign_source_account_key
    ON access_entries (access_review_campaign_id, access_source_id, account_key);

-- 5. access_review_campaign_source_fetches: tracks per-source fetch lifecycle
CREATE TABLE access_review_campaign_source_fetches (
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT NOT NULL REFERENCES access_sources(id) ON DELETE CASCADE,
    status access_review_campaign_source_fetch_status NOT NULL,
    fetched_accounts_count INTEGER NOT NULL,
    attempt_count INTEGER NOT NULL,
    last_error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (access_review_campaign_id, access_source_id)
);
