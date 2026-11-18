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
