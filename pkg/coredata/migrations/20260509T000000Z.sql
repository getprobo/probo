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
