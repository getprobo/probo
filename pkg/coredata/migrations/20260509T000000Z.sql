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

-- Split SCRIPT/IFRAME tracking out of tracker_patterns/detected_trackers into
-- its own tracker_resources table keyed by (banner, type, origin, path).
-- SCRIPT/IFRAME data only landed last week and is not in the production
-- database yet, so existing rows are dropped rather than backfilled.

CREATE TYPE tracker_resource_type AS ENUM ('SCRIPT', 'IFRAME');

CREATE TABLE tracker_resources (
    id                 TEXT NOT NULL PRIMARY KEY,
    tenant_id          TEXT NOT NULL,
    organization_id    TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cookie_banner_id   TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    cookie_category_id TEXT NOT NULL REFERENCES cookie_categories(id) ON DELETE CASCADE,
    resource_type      tracker_resource_type NOT NULL,
    origin             TEXT NOT NULL,
    path               TEXT NOT NULL,
    display_name       TEXT NOT NULL,
    description        TEXT NOT NULL,
    excluded           BOOLEAN NOT NULL,
    last_detected_at   TIMESTAMP WITH TIME ZONE,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_tracker_resources_unique_resource_per_banner
    ON tracker_resources (cookie_banner_id, resource_type, origin, path);

-- Drop SCRIPT/IFRAME rows from the legacy tables before recreating the enum.
DELETE FROM detected_trackers WHERE tracker_type IN ('SCRIPT', 'IFRAME');
DELETE FROM tracker_patterns  WHERE tracker_type IN ('SCRIPT', 'IFRAME');

-- Recreate tracker_type without SCRIPT/IFRAME. Postgres has no
-- "remove enum value" so we swap the type via a _new alias.
CREATE TYPE tracker_type_new AS ENUM (
    'COOKIE', 'LOCAL_STORAGE', 'SESSION_STORAGE', 'INDEXED_DB'
);

ALTER TABLE tracker_patterns
    ALTER COLUMN tracker_type TYPE tracker_type_new
        USING tracker_type::text::tracker_type_new;

ALTER TABLE detected_trackers
    ALTER COLUMN tracker_type TYPE tracker_type_new
        USING tracker_type::text::tracker_type_new;

DROP TYPE tracker_type;
ALTER TYPE tracker_type_new RENAME TO tracker_type;
