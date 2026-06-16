-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
