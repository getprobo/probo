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

-- Per-source inventory sync schedule (AR-P1-08). NULL next_sync_at means due
-- immediately. Cadence enqueue skips sources that already have a QUEUED or
-- RUNNING run (see LoadNextDueForInventorySyncSkipLocked).

ALTER TABLE access_review_sources
    ADD COLUMN last_successful_sync_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN next_sync_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN sync_consecutive_failures INTEGER NOT NULL DEFAULT 0;

ALTER TABLE access_review_sources
    ALTER COLUMN sync_consecutive_failures DROP DEFAULT;

-- Scheduler claim orders by next_sync_at NULLS FIRST across tenants (same
-- shape as iam_scim_bridges next_sync indexing).
CREATE INDEX access_review_sources_next_sync_at_idx
    ON access_review_sources (next_sync_at ASC NULLS FIRST);
