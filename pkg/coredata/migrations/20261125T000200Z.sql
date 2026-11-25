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

-- Campaign→inventory backfill looks up the latest SUCCESS fetch per live
-- source (and lists orgs with any such fetch) on every service boot. Neither
-- access_review_source_id nor SUCCESS fetch attempts had a supporting index.

CREATE INDEX idx_access_review_campaign_sources_access_review_source_id
    ON access_review_campaign_sources (access_review_source_id)
    WHERE access_review_source_id IS NOT NULL;

CREATE INDEX idx_access_review_campaign_source_fetch_attempts_success
    ON access_review_campaign_source_fetch_attempts (
        access_review_campaign_source_id,
        (COALESCE(completed_at, created_at)) DESC,
        attempt_number DESC
    )
    WHERE status = 'SUCCESS';
