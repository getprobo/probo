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

ALTER TABLE files ADD COLUMN visibility TEXT NOT NULL DEFAULT 'PRIVATE';

-- Backfill trust center logos to PUBLIC
UPDATE files SET visibility = 'PUBLIC'
WHERE id IN (
    SELECT logo_file_id FROM trust_centers WHERE logo_file_id IS NOT NULL
    UNION
    SELECT dark_logo_file_id FROM trust_centers WHERE dark_logo_file_id IS NOT NULL
);

-- Backfill organization logos to PUBLIC
UPDATE files SET visibility = 'PUBLIC'
WHERE id IN (
    SELECT logo_file_id FROM organizations WHERE logo_file_id IS NOT NULL
    UNION
    SELECT horizontal_logo_file_id FROM organizations WHERE horizontal_logo_file_id IS NOT NULL
);

-- Backfill framework logos to PUBLIC
UPDATE files SET visibility = 'PUBLIC'
WHERE id IN (
    SELECT light_logo_file_id FROM frameworks WHERE light_logo_file_id IS NOT NULL
    UNION
    SELECT dark_logo_file_id FROM frameworks WHERE dark_logo_file_id IS NOT NULL
);

ALTER TABLE files ALTER COLUMN visibility DROP DEFAULT;
