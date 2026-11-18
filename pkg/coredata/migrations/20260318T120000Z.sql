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
