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
        USING status::text::access_review_campaign_status;

DROP TYPE access_review_campaign_status_old;
