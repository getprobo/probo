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

-- Snapshot CSV payloads on campaign sources (20260611T010000Z shipped without
-- this column; fresh installs get it from the CREATE TABLE in that migration).

ALTER TABLE access_review_campaign_sources
    ADD COLUMN IF NOT EXISTS csv_data TEXT;

UPDATE access_review_campaign_sources cs
SET csv_data = s.csv_data
FROM access_review_sources s
WHERE cs.access_review_source_id = s.id
  AND cs.csv_data IS NULL
  AND s.csv_data IS NOT NULL;
