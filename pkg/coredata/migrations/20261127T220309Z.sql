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

-- A publish copies the active assets at that moment. The copy is inserted
-- and never updated. asset_id and the link id arrays are not foreign keys:
-- a later edit, sync, deactivation, or delete of the live row leaves the
-- copy alone. Deleting the snapshot removes its copies.

CREATE TABLE asset_snapshots (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    captured_at TIMESTAMP WITH TIME ZONE NOT NULL,
    actor_profile_id TEXT REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    name TEXT,
    note TEXT
);

CREATE TABLE asset_snapshot_assets (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    snapshot_id TEXT NOT NULL REFERENCES asset_snapshots(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    organization_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    name TEXT NOT NULL,
    amount INTEGER NOT NULL,
    owner_profile_id TEXT,
    asset_type asset_type NOT NULL,
    data_types_stored TEXT,
    kind TEXT,
    external_id TEXT,
    region TEXT,
    configuration JSONB,
    schema_version INTEGER,
    last_seen_at TIMESTAMP WITH TIME ZONE,
    status asset_status NOT NULL,
    connector_id TEXT,
    asset_created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    asset_updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    third_party_ids TEXT[] NOT NULL,
    business_function_ids TEXT[] NOT NULL,
    tag_ids TEXT[] NOT NULL
);
