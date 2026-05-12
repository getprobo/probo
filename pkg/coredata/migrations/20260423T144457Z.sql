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
