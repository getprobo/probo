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

-- The common-pattern enrichment worker fills descriptions on
-- common_tracker_patterns using an agent with web search, then fans the
-- result out to every linked tracker pattern. enrichment_requested_at is
-- the work queue (claimed FOR UPDATE SKIP LOCKED); enriched_at marks a
-- row as terminally enriched so it is never re-enqueued.
ALTER TABLE common_tracker_patterns
    ADD COLUMN enrichment_requested_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN enriched_at             TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_common_tracker_patterns_enrichment
    ON common_tracker_patterns (enrichment_requested_at)
    WHERE enrichment_requested_at IS NOT NULL;

-- Enqueue existing description-less rows for a first enrichment pass.
UPDATE common_tracker_patterns
SET enrichment_requested_at = NOW()
WHERE description = '';
