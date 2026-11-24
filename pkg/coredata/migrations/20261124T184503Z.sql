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

-- Live inventory source sync runs (AR-P1-03). Claim/stale helpers use
-- FOR UPDATE SKIP LOCKED. Change events may reference a completed run.

CREATE TYPE access_source_sync_run_status AS ENUM (
    'QUEUED',
    'RUNNING',
    'SUCCESS',
    'FAILED'
);

CREATE TABLE access_source_sync_runs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    access_review_source_id TEXT NOT NULL REFERENCES access_review_sources(id) ON UPDATE CASCADE ON DELETE CASCADE,
    status access_source_sync_run_status NOT NULL,
    grants_upserted_count INTEGER NOT NULL,
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

ALTER TABLE access_inventory_change_events
    ADD CONSTRAINT access_inventory_change_events_sync_run_id_fkey
    FOREIGN KEY (sync_run_id)
    REFERENCES access_source_sync_runs(id)
    ON UPDATE CASCADE
    ON DELETE SET NULL;
