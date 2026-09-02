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

-- Move the source-name retry policy into the claim predicate.
-- The DEFAULT is kept rather than dropped in this migration: probod deploys
-- blue/green, so pods running the previous release keep inserting sources
-- without this column until the rollout finishes, and those writes need it.
-- Dropping it belongs in a later migration.
ALTER TABLE access_review_sources
    ADD COLUMN name_sync_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN name_sync_next_attempt_at TIMESTAMP WITH TIME ZONE;

-- Leads on created_at so the claim walks the partial set in order and stops at
-- the first row. The backoff is an OR range, which cannot be walked in order,
-- so it stays a recheck against the handful of rows the predicate admits.
CREATE INDEX idx_access_review_sources_name_sync
    ON access_review_sources (created_at)
    WHERE connector_id IS NOT NULL AND name_synced_at IS NULL;
