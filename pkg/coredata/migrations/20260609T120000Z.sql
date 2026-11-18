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

ALTER TABLE
    common_third_parties
ADD
    COLUMN enrichment_requested_at TIMESTAMP WITH TIME ZONE,
ADD
    COLUMN enrichment JSONB,
ADD
    COLUMN enrichment_attempts INTEGER NOT NULL DEFAULT 0;

-- The DEFAULT only backfills existing rows; drop it so inserts must
-- supply the value explicitly.
ALTER TABLE
    common_third_parties
ALTER COLUMN
    enrichment_attempts DROP DEFAULT;

-- Partial index backs the enrichment worker's claim query, which polls
-- only rows currently queued for enrichment.
CREATE INDEX common_third_parties_enrichment_requested_at_idx
    ON common_third_parties (enrichment_requested_at)
    WHERE enrichment_requested_at IS NOT NULL;
