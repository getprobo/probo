-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

CREATE TYPE cookie_pattern_match_type AS ENUM ('EXACT', 'PREFIX');

CREATE TABLE cookie_patterns (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    cookie_category_id TEXT NOT NULL REFERENCES cookie_categories(id) ON DELETE CASCADE,
    pattern TEXT NOT NULL,
    match_type cookie_pattern_match_type NOT NULL,
    display_name TEXT NOT NULL,
    duration TEXT NOT NULL,
    description TEXT NOT NULL,
    source cookie_source NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_cookie_patterns_unique_pattern_per_banner
    ON cookie_patterns (cookie_banner_id, pattern);

-- Backfill: create an EXACT pattern for each existing cookie
INSERT INTO cookie_patterns (
    id, tenant_id, organization_id, cookie_banner_id,
    cookie_category_id, pattern, match_type, display_name,
    duration, description, source, created_at, updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(c.tenant_id), 88),
    c.tenant_id, c.organization_id, c.cookie_banner_id,
    c.cookie_category_id, c.name, 'EXACT', c.name,
    c.duration, c.description, c.source, c.created_at, c.updated_at
FROM cookies c;

-- Link each cookie to its pattern
ALTER TABLE cookies ADD COLUMN cookie_pattern_id TEXT REFERENCES cookie_patterns(id) ON DELETE CASCADE;

UPDATE cookies c
SET cookie_pattern_id = cp.id
FROM cookie_patterns cp
WHERE cp.cookie_banner_id = c.cookie_banner_id
  AND cp.pattern = c.name
  AND cp.match_type = 'EXACT';

ALTER TABLE cookies ALTER COLUMN cookie_pattern_id SET NOT NULL;

-- Category and description now live on the pattern
ALTER TABLE cookies DROP COLUMN cookie_category_id;
ALTER TABLE cookies DROP COLUMN description;
