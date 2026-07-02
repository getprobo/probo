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
