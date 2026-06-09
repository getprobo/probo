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
