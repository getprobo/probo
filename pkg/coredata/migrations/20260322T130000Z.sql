-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
    ADD COLUMN cookies_v2 JSONB NOT NULL DEFAULT '[]';

UPDATE cookie_categories
SET cookies_v2 = (
    SELECT COALESCE(jsonb_agg(jsonb_build_object('name', c, 'duration', '', 'description', '')), '[]'::jsonb)
    FROM unnest(cookies) AS c
)
WHERE array_length(cookies, 1) > 0;

ALTER TABLE cookie_categories DROP COLUMN cookies;
ALTER TABLE cookie_categories RENAME COLUMN cookies_v2 TO cookies;
