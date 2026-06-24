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

-- Add organization_id to campaign source snapshots and fetch attempts to
-- avoid JOINs in AuthorizationAttributes lookups.

-- 1. access_review_campaign_sources
ALTER TABLE access_review_campaign_sources
    ADD COLUMN organization_id TEXT REFERENCES organizations(id);

UPDATE access_review_campaign_sources cs
SET organization_id = c.organization_id
FROM access_review_campaigns c
WHERE cs.access_review_campaign_id = c.id;

ALTER TABLE access_review_campaign_sources
    ALTER COLUMN organization_id SET NOT NULL;

-- 2. access_review_campaign_source_fetch_attempts
ALTER TABLE access_review_campaign_source_fetch_attempts
    ADD COLUMN organization_id TEXT REFERENCES organizations(id);

UPDATE access_review_campaign_source_fetch_attempts fa
SET organization_id = cs.organization_id
FROM access_review_campaign_sources cs
WHERE fa.access_review_campaign_source_id = cs.id;

ALTER TABLE access_review_campaign_source_fetch_attempts
    ALTER COLUMN organization_id SET NOT NULL;
