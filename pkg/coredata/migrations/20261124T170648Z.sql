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

-- Live access inventory: principal x resource x grant (AR-P1-01).
-- Reviewer decisions/flags stay on campaign entries; these rows are sync state only.

CREATE TYPE access_inventory_grant_presence AS ENUM (
    'PRESENT',
    'DISAPPEARED'
);

CREATE TABLE access_inventory_principals (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    access_review_source_id TEXT NOT NULL REFERENCES access_review_sources(id) ON UPDATE CASCADE ON DELETE CASCADE,
    principal_key TEXT NOT NULL,
    email TEXT NOT NULL,
    full_name TEXT NOT NULL,
    external_id TEXT NOT NULL,
    job_title TEXT NOT NULL,
    is_admin BOOLEAN,
    mfa_status mfa_status NOT NULL,
    auth_method auth_method NOT NULL,
    account_type TEXT NOT NULL,
    active BOOLEAN,
    last_login TIMESTAMP WITH TIME ZONE,
    account_created_at TIMESTAMP WITH TIME ZONE,
    first_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_source_id, principal_key),
    UNIQUE (access_review_source_id, id, principal_key)
);

CREATE TABLE access_inventory_resources (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    access_review_source_id TEXT NOT NULL REFERENCES access_review_sources(id) ON UPDATE CASCADE ON DELETE CASCADE,
    resource_key TEXT NOT NULL,
    display_name TEXT NOT NULL,
    kind TEXT NOT NULL,
    first_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_source_id, resource_key),
    UNIQUE (access_review_source_id, id, resource_key)
);

CREATE TABLE access_inventory_grants (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    access_review_source_id TEXT NOT NULL REFERENCES access_review_sources(id) ON UPDATE CASCADE ON DELETE CASCADE,
    principal_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    principal_key TEXT NOT NULL,
    resource_key TEXT NOT NULL,
    entitlement_key TEXT NOT NULL,
    entitlement_display_name TEXT NOT NULL,
    presence access_inventory_grant_presence NOT NULL,
    consecutive_missing_syncs INTEGER NOT NULL,
    first_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    provider_changed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_source_id, principal_key, resource_key, entitlement_key),
    FOREIGN KEY (access_review_source_id, principal_id, principal_key)
        REFERENCES access_inventory_principals (access_review_source_id, id, principal_key)
        ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (access_review_source_id, resource_id, resource_key)
        REFERENCES access_inventory_resources (access_review_source_id, id, resource_key)
        ON UPDATE CASCADE ON DELETE CASCADE
);
