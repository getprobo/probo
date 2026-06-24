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

ALTER TABLE cookie_categories
    ADD COLUMN slug TEXT NOT NULL DEFAULT '';

UPDATE cookie_categories
SET slug = COALESCE(
    NULLIF(LOWER(REGEXP_REPLACE(REGEXP_REPLACE(name, '[^a-zA-Z0-9]+', '-', 'g'), '^-|-$', '', 'g')), ''),
    'category-' || SUBSTR(id, 1, 8)
);

WITH dupes AS (
    SELECT id,
           cookie_banner_id,
           slug,
           ROW_NUMBER() OVER (PARTITION BY cookie_banner_id, slug ORDER BY created_at) AS rn
    FROM cookie_categories
)
UPDATE cookie_categories
SET slug = cookie_categories.slug || '-' || dupes.rn
FROM dupes
WHERE cookie_categories.id = dupes.id AND dupes.rn > 1;

ALTER TABLE cookie_categories ALTER COLUMN slug DROP DEFAULT;

CREATE UNIQUE INDEX idx_cookie_categories_unique_slug_per_banner
    ON cookie_categories (cookie_banner_id, slug);
