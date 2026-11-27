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

-- Append-only history of an asset row. Rows are inserted and never updated.
-- This migration does not backfill events for assets that already exist.
-- asset_sync_id is not a foreign key: deleting a sync must not update or
-- remove the event.

CREATE TYPE asset_change_type AS ENUM (
    'CREATED',
    'UPDATED',
    'DEACTIVATED'
);

CREATE TABLE asset_change_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES assets(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    observed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    asset_sync_id TEXT,
    actor_profile_id TEXT REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    change_type asset_change_type NOT NULL,
    before JSONB,
    after JSONB
);
