-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

-- Access review: decouple campaign data from live sources.
--
-- Introduces a per-campaign source snapshot (access_review_campaign_sources)
-- that owns the source identity (name/category/connector) so a review survives
-- the deletion of the live access source. Fetch tracking becomes an append-only
-- log (access_review_campaign_source_fetch_attempts) so each fetch run keeps its own
-- error. Access entries are repointed from the live source to the snapshot.

-- Helper to mint a valid GID (base64url of: tenant 8B | entity type 2B |
-- timestamp 8B | random 6B) for back-filled rows.
CREATE FUNCTION pg_temp.gen_gid(tenant_text text, entity_type int) RETURNS text AS $$
    SELECT translate(
        encode(
            decode(rpad(translate(tenant_text, '-_', '+/'), 12, '='), 'base64')
            || set_byte(set_byte('\x0000'::bytea, 0, (entity_type >> 8) & 255), 1, entity_type & 255)
            || int8send((extract(epoch FROM clock_timestamp()) * 1000)::bigint)
            || substring(uuid_send(gen_random_uuid()) FROM 1 FOR 6),
            'base64'),
        '+/', '-_')
$$ LANGUAGE sql VOLATILE;

-- 1. Per-campaign source snapshot. The live access source link is nullable and
-- ON DELETE SET NULL so deleting a source preserves the review's snapshot.
CREATE TABLE access_review_campaign_sources (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT REFERENCES access_sources(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    category access_source_category NOT NULL,
    connector_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_campaign_id, access_source_id)
);

-- 2. Append-only fetch attempts. Each run is a new row; the snapshot's current
-- state is the latest attempt.
CREATE TABLE access_review_campaign_source_fetch_attempts (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    access_review_campaign_source_id TEXT NOT NULL REFERENCES access_review_campaign_sources(id) ON DELETE CASCADE,
    attempt_number INTEGER NOT NULL,
    status access_review_campaign_source_fetch_status NOT NULL,
    fetched_accounts_count INTEGER NOT NULL,
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_campaign_source_id, attempt_number)
);

CREATE INDEX idx_fetch_attempts_queued
    ON access_review_campaign_source_fetch_attempts (created_at)
    WHERE status = 'QUEUED';

-- 3. Back-fill snapshots from every (campaign, source) pair already present in
-- the scope, entries, or fetches.
INSERT INTO access_review_campaign_sources (
    id, tenant_id, organization_id, access_review_campaign_id, access_source_id,
    name, category, connector_id, created_at, updated_at
)
SELECT
    pg_temp.gen_gid(s.tenant_id, 102),
    s.tenant_id,
    s.organization_id,
    p.access_review_campaign_id,
    p.access_source_id,
    s.name,
    s.category,
    s.connector_id,
    now(),
    now()
FROM (
    SELECT access_review_campaign_id, access_source_id
    FROM access_review_campaign_scope_systems
    UNION
    SELECT access_review_campaign_id, access_source_id
    FROM access_entries
    UNION
    SELECT access_review_campaign_id, access_source_id
    FROM access_review_campaign_source_fetches
) p
JOIN access_sources s ON s.id = p.access_source_id;

-- 4. Repoint access entries at the snapshot.
ALTER TABLE access_entries
    ADD COLUMN access_review_campaign_source_id TEXT;

UPDATE access_entries e
SET access_review_campaign_source_id = cs.id
FROM access_review_campaign_sources cs
WHERE cs.access_review_campaign_id = e.access_review_campaign_id
  AND cs.access_source_id = e.access_source_id;

ALTER TABLE access_entries
    ALTER COLUMN access_review_campaign_source_id SET NOT NULL;

ALTER TABLE access_entries
    ADD CONSTRAINT access_entries_campaign_source_id_fkey
    FOREIGN KEY (access_review_campaign_source_id)
    REFERENCES access_review_campaign_sources(id) ON DELETE CASCADE;

DROP INDEX idx_access_entries_campaign_source_account_key;

CREATE UNIQUE INDEX idx_access_entries_campaign_source_account_key
    ON access_entries (access_review_campaign_source_id, account_key);

ALTER TABLE access_entries
    DROP COLUMN access_source_id;

-- 5. Migrate fetch rows into the append-only attempt log (one attempt each).
INSERT INTO access_review_campaign_source_fetch_attempts (
    id, tenant_id, organization_id, access_review_campaign_source_id, attempt_number,
    status, fetched_accounts_count, error, started_at, completed_at,
    created_at, updated_at
)
SELECT
    pg_temp.gen_gid(f.tenant_id, 103),
    f.tenant_id,
    cs.organization_id,
    cs.id,
    GREATEST(f.attempt_count, 1),
    f.status,
    f.fetched_accounts_count,
    f.last_error,
    f.started_at,
    f.completed_at,
    f.created_at,
    f.updated_at
FROM access_review_campaign_source_fetches f
JOIN access_review_campaign_sources cs
    ON cs.access_review_campaign_id = f.access_review_campaign_id
   AND cs.access_source_id = f.access_source_id;

-- 6. Drop the superseded tables.
DROP TABLE access_review_campaign_source_fetches;
DROP TABLE access_review_campaign_scope_systems;
