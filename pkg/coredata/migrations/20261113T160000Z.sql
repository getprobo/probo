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

-- Source promotion is queued on the banner that wrote the detections.
-- requested_at is the start of the window the next run reads. since keeps
-- that start after the claim, so a retry still sees the same rows. The
-- default backfills attempts, then goes away so later inserts set the count.
ALTER TABLE cookie_banners
    ADD COLUMN source_promotion_requested_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN source_promotion_since TIMESTAMP WITH TIME ZONE,
    ADD COLUMN source_promotion_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN source_promotion_attempted_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE cookie_banners
    ALTER COLUMN source_promotion_attempts DROP DEFAULT;

-- One run covers every detection a grouped banner already has. Corsless
-- banners stay out of promotion.
UPDATE cookie_banners
SET source_promotion_requested_at = created_at
WHERE cmp_domain_id IS NOT NULL
  AND NOT (capabilities @> '{"corsless": true}'::jsonb);

ALTER TABLE cmp_domains
    DROP COLUMN source_promotion_requested_at,
    DROP COLUMN source_promotion_attempts,
    DROP COLUMN source_promotion_attempted_at;
