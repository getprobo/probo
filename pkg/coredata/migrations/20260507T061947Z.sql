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

-- Drop the FAILED status from access_review_campaign_status. A campaign is
-- no longer marked as failed when individual sources fail to fetch: the
-- failure stays surfaced on the source fetch (status + last_error) and the
-- review can still be performed on the sources that succeeded.

UPDATE access_review_campaigns
SET status = 'PENDING_ACTIONS'
WHERE status = 'FAILED';

ALTER TYPE access_review_campaign_status RENAME TO access_review_campaign_status_old;

CREATE TYPE access_review_campaign_status AS ENUM (
    'DRAFT',
    'IN_PROGRESS',
    'PENDING_ACTIONS',
    'COMPLETED',
    'CANCELLED'
);

ALTER TABLE access_review_campaigns
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN status TYPE access_review_campaign_status
        USING status::text::access_review_campaign_status,
    ALTER COLUMN status SET DEFAULT 'DRAFT';

DROP TYPE access_review_campaign_status_old;
