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

CREATE TYPE cookie_category_kind AS ENUM ('NORMAL', 'NECESSARY', 'UNCATEGORISED');

ALTER TABLE cookie_categories ADD COLUMN kind cookie_category_kind NOT NULL DEFAULT 'NORMAL';
UPDATE cookie_categories SET kind = 'NECESSARY' WHERE required = TRUE;
ALTER TABLE cookie_categories ALTER COLUMN kind DROP DEFAULT;

DROP INDEX idx_cookie_categories_one_required_per_banner;
ALTER TABLE cookie_categories DROP COLUMN required;

CREATE UNIQUE INDEX idx_cookie_categories_one_necessary_per_banner
    ON cookie_categories (cookie_banner_id) WHERE kind = 'NECESSARY';
CREATE UNIQUE INDEX idx_cookie_categories_one_uncategorised_per_banner
    ON cookie_categories (cookie_banner_id) WHERE kind = 'UNCATEGORISED';
