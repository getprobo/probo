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
