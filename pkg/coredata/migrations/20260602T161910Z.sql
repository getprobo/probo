-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

ALTER TABLE cookie_banners
    ADD COLUMN policy_document_id TEXT,
    ADD COLUMN policy_generation_requested_at TIMESTAMP WITH TIME ZONE;

-- Backfill: any banner that already has a published version gets its tracker
-- policy generated on the worker's first pass.
UPDATE cookie_banners
SET policy_generation_requested_at = NOW()
WHERE id IN (
    SELECT DISTINCT cookie_banner_id
    FROM cookie_banner_versions
    WHERE state = 'PUBLISHED'
);