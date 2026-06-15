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

-- Unify the enrichment-tracking model across common_tracker_patterns and
-- common_third_parties. common_tracker_patterns gains a JSONB provenance
-- payload and an attempt counter (mirroring common_third_parties), and its
-- enriched_at done-flag is renamed to last_enrichment_attempt_at, stamped at
-- claim time. "Enriched" is now detected via the enrichment payload presence,
-- not the timestamp.
ALTER TABLE common_tracker_patterns
    ADD COLUMN enrichment          JSONB,
    ADD COLUMN enrichment_attempts INTEGER NOT NULL DEFAULT 0;

-- The DEFAULT only backfills existing rows; drop it so inserts must supply
-- the value explicitly.
ALTER TABLE common_tracker_patterns
    ALTER COLUMN enrichment_attempts DROP DEFAULT;

ALTER TABLE common_tracker_patterns
    RENAME COLUMN enriched_at TO last_enrichment_attempt_at;

-- common_third_parties already carries enrichment/enrichment_attempts; give
-- it the same explicit last-attempt timestamp (previously only recorded in
-- the enrichment JSON) so the stale-recovery clock has a dedicated column.
ALTER TABLE common_third_parties
    ADD COLUMN last_enrichment_attempt_at TIMESTAMP WITH TIME ZONE;
