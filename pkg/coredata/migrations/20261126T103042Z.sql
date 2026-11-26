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

-- Discovery columns live on the existing assets row. A typed row leaves
-- them null. connector_id is not a foreign key: deleting the connector
-- leaves the asset, including this id.

CREATE TYPE asset_status AS ENUM (
    'ACTIVE',
    'DEACTIVATED'
);

CREATE TYPE asset_sync_status AS ENUM (
    'RUNNING',
    'SUCCEEDED',
    'FAILED'
);

ALTER TABLE assets
    ALTER COLUMN owner_profile_id DROP NOT NULL,
    ALTER COLUMN data_types_stored DROP NOT NULL,
    ADD COLUMN kind TEXT,
    ADD COLUMN external_id TEXT,
    ADD COLUMN region TEXT,
    ADD COLUMN configuration JSONB,
    ADD COLUMN schema_version INTEGER,
    ADD COLUMN last_seen_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN connector_id TEXT,
    -- status stays NOT NULL DEFAULT 'ACTIVE'. probod deploys blue/green,
    -- and pods still running the previous release insert assets without
    -- this column until the rollout finishes.
    ADD COLUMN status asset_status NOT NULL DEFAULT 'ACTIVE';

CREATE UNIQUE INDEX assets_discovered_identity_idx
    ON assets (organization_id, connector_id, kind, external_id)
    WHERE connector_id IS NOT NULL
        AND kind IS NOT NULL
        AND external_id IS NOT NULL;

CREATE TABLE asset_syncs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    connector_account_id TEXT NOT NULL REFERENCES connector_accounts(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    status asset_sync_status NOT NULL,
    error TEXT,
    completed_kinds TEXT[] NOT NULL,
    attempt_number INTEGER NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX asset_syncs_connector_account_created_at_idx
    ON asset_syncs (connector_account_id, created_at);
