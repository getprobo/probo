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

CREATE TYPE tracker_type AS ENUM (
    'COOKIE', 'LOCAL_STORAGE', 'SESSION_STORAGE', 'INDEXED_DB', 'SCRIPT', 'IFRAME'
);

CREATE TABLE tracker_patterns (
    id                 TEXT NOT NULL PRIMARY KEY,
    tenant_id          TEXT NOT NULL,
    organization_id    TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cookie_banner_id   TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    cookie_category_id TEXT NOT NULL REFERENCES cookie_categories(id) ON DELETE CASCADE,
    tracker_type       tracker_type NOT NULL,
    pattern            TEXT NOT NULL,
    match_type         cookie_pattern_match_type NOT NULL,
    display_name       TEXT NOT NULL,
    description        TEXT NOT NULL,
    excluded           BOOLEAN NOT NULL,
    max_age_seconds    INTEGER,
    source             cookie_source,
    last_matched_at    TIMESTAMP WITH TIME ZONE,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_tracker_patterns_unique_pattern_per_banner
    ON tracker_patterns (cookie_banner_id, tracker_type, pattern);

CREATE TABLE detected_trackers (
    id                 TEXT NOT NULL PRIMARY KEY,
    tenant_id          TEXT NOT NULL,
    cookie_banner_id   TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    tracker_pattern_id TEXT REFERENCES tracker_patterns(id) ON DELETE CASCADE,
    tracker_type       tracker_type NOT NULL,
    identifier         TEXT NOT NULL,
    max_age_seconds    INTEGER,
    source             cookie_source,
    value_size         INTEGER,
    last_detected_at   TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_detected_trackers_unique_identifier_per_banner
    ON detected_trackers (cookie_banner_id, tracker_type, identifier);

-- Backfill: copy cookie_patterns into tracker_patterns
INSERT INTO tracker_patterns (
    id, tenant_id, organization_id, cookie_banner_id, cookie_category_id,
    tracker_type, pattern, match_type, display_name, description,
    excluded, max_age_seconds, source, last_matched_at, created_at, updated_at
)
SELECT
    id, tenant_id, organization_id, cookie_banner_id, cookie_category_id,
    'COOKIE', pattern, match_type, display_name, description,
    excluded, max_age_seconds, source, last_matched_at, created_at, updated_at
FROM cookie_patterns;

-- Backfill: copy cookies into detected_trackers
INSERT INTO detected_trackers (
    id, tenant_id, cookie_banner_id, tracker_pattern_id,
    tracker_type, identifier, max_age_seconds, source,
    last_detected_at, created_at, updated_at
)
SELECT
    id, tenant_id, cookie_banner_id, cookie_pattern_id,
    'COOKIE', name, max_age_seconds, source,
    last_detected_at, created_at, updated_at
FROM cookies;
