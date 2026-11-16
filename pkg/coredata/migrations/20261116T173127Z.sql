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

-- Asset inventory. The Go type on assets is InventoryAsset, because
-- Asset is the register. The register stays in assets_registry.
-- Discovery writes assets only.
--
-- Entity types: 148 asset source, 149 inventory asset, 150 asset sync.
-- A source is one external account. Deleting the connector account
-- clears connector_account_id and leaves the source and its assets.

CREATE TYPE assets_sync_status AS ENUM (
    'QUEUED',
    'RUNNING',
    'SUCCEEDED',
    'FAILED'
);

CREATE TABLE assets_sources (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    provider connector_provider NOT NULL,
    external_account_id TEXT NOT NULL,
    connector_account_id TEXT REFERENCES connector_accounts(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    next_sync_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT assets_sources_identity_key
        UNIQUE (organization_id, provider, external_account_id)
);

CREATE INDEX assets_sources_due_idx
    ON assets_sources (next_sync_at)
    WHERE connector_account_id IS NOT NULL;

CREATE TABLE assets (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    assets_source_id TEXT NOT NULL REFERENCES assets_sources(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    kind TEXT NOT NULL,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    region TEXT,
    configuration JSONB NOT NULL,
    schema_version INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT assets_identity_key
        UNIQUE (assets_source_id, kind, external_id)
);

CREATE TABLE assets_syncs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    assets_source_id TEXT NOT NULL REFERENCES assets_sources(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    status assets_sync_status NOT NULL,
    error TEXT,
    completed_kinds TEXT[] NOT NULL,
    attempt_number INTEGER NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX assets_syncs_source_created_at_idx
    ON assets_syncs (assets_source_id, created_at);

CREATE INDEX assets_syncs_running_started_at_idx
    ON assets_syncs (started_at)
    WHERE status = 'RUNNING';

-- AWS is the only provider with a discoverer. A later provider needs
-- its own backfill. The newest account wins when one external id has
-- more than one row.
INSERT INTO assets_sources (
    id,
    tenant_id,
    organization_id,
    provider,
    external_account_id,
    connector_account_id,
    next_sync_at,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(picked.tenant_id), 148),
    picked.tenant_id,
    picked.organization_id,
    picked.provider,
    picked.external_account_id,
    picked.connector_account_id,
    NOW(),
    NOW(),
    NOW()
FROM (
    SELECT DISTINCT ON (a.organization_id, c.provider, a.external_account_id)
        a.tenant_id,
        a.organization_id,
        c.provider,
        a.external_account_id,
        a.id AS connector_account_id
    FROM
        connector_accounts a
    INNER JOIN
        connectors c ON c.id = a.connector_id
    WHERE
        c.provider = 'AWS'
    ORDER BY
        a.organization_id,
        c.provider,
        a.external_account_id,
        a.updated_at DESC
) picked
ON CONFLICT ON CONSTRAINT assets_sources_identity_key DO UPDATE
SET
    connector_account_id = EXCLUDED.connector_account_id,
    next_sync_at = EXCLUDED.next_sync_at,
    updated_at = EXCLUDED.updated_at;
