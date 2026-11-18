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

-- Give the global tracker-pattern catalog a terminal attribution verdict so a
-- row can record that it has no third party (a first-party or generic
-- artifact) rather than only "linked" vs "not yet linked". UNDETERMINED rows
-- are still probed by the mapping pipeline; THIRD_PARTY rows carry a vendor;
-- FIRST_PARTY rows are terminal and the pipeline never attributes them again.
CREATE TYPE common_tracker_pattern_attribution AS ENUM (
    'UNDETERMINED',
    'THIRD_PARTY',
    'FIRST_PARTY'
);

ALTER TABLE common_tracker_patterns
    ADD COLUMN attribution common_tracker_pattern_attribution NOT NULL DEFAULT 'UNDETERMINED';

-- Backfill: any row already carrying a vendor is, by definition, attributed to
-- a third party. The DEFAULT covers the rest (UNDETERMINED).
UPDATE common_tracker_patterns
SET attribution = 'THIRD_PARTY'
WHERE common_third_party_id IS NOT NULL;

-- The DEFAULT only backfills existing rows; drop it so inserts must supply the
-- value explicitly, matching the cookie_source convention.
ALTER TABLE common_tracker_patterns
    ALTER COLUMN attribution DROP DEFAULT;
