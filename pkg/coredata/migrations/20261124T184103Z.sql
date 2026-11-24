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

-- Live access inventory change events (AR-P1-02). Append-only; sync_run_id is
-- nullable until AccessSourceSyncRun lands in AR-P1-03.

CREATE TYPE access_inventory_change_type AS ENUM (
    'APPEARED',
    'DISAPPEARED',
    'ATTRIBUTE_CHANGED'
);

CREATE TABLE access_inventory_change_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    access_review_source_id TEXT NOT NULL REFERENCES access_review_sources(id) ON UPDATE CASCADE ON DELETE CASCADE,
    grant_id TEXT REFERENCES access_inventory_grants(id) ON UPDATE CASCADE ON DELETE SET NULL,
    principal_key TEXT NOT NULL,
    resource_key TEXT NOT NULL,
    entitlement_key TEXT NOT NULL,
    change_type access_inventory_change_type NOT NULL,
    observed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    provider_occurred_at TIMESTAMP WITH TIME ZONE,
    sync_run_id TEXT,
    before JSONB,
    after JSONB,
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_source_id, idempotency_key),
    CHECK (
        (change_type = 'APPEARED' AND before IS NULL AND after IS NOT NULL)
        OR (change_type = 'DISAPPEARED' AND before IS NOT NULL AND after IS NULL)
        OR (change_type = 'ATTRIBUTE_CHANGED' AND before IS NOT NULL AND after IS NOT NULL)
    )
);
