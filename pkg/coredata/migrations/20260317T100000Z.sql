-- Remove the access_reviews singleton table. Access sources and campaigns now
-- reference organizations directly.

-- 1. Add organization_id to access_sources (backfill from access_reviews).
ALTER TABLE access_sources ADD COLUMN organization_id TEXT;

UPDATE access_sources
SET organization_id = ar.organization_id
FROM access_reviews ar
WHERE access_sources.access_review_id = ar.id;

ALTER TABLE access_sources ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE access_sources
    ADD CONSTRAINT fk_access_sources_organization
    FOREIGN KEY (organization_id) REFERENCES organizations(id);

CREATE INDEX idx_access_sources_organization_id ON access_sources(organization_id);

-- 2. Add organization_id to access_review_campaigns (backfill from access_reviews).
ALTER TABLE access_review_campaigns ADD COLUMN organization_id TEXT;

UPDATE access_review_campaigns
SET organization_id = ar.organization_id
FROM access_reviews ar
WHERE access_review_campaigns.access_review_id = ar.id;

ALTER TABLE access_review_campaigns ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE access_review_campaigns
    ADD CONSTRAINT fk_access_review_campaigns_organization
    FOREIGN KEY (organization_id) REFERENCES organizations(id);

CREATE INDEX idx_access_review_campaigns_organization_id ON access_review_campaigns(organization_id);

-- 3. Move identity_source_id to organizations.
ALTER TABLE organizations ADD COLUMN identity_source_id TEXT;

UPDATE organizations
SET identity_source_id = ar.identity_source_id
FROM access_reviews ar
WHERE organizations.id = ar.organization_id
  AND ar.identity_source_id IS NOT NULL;

ALTER TABLE organizations
    ADD CONSTRAINT fk_organizations_identity_source
    FOREIGN KEY (identity_source_id) REFERENCES access_sources(id) ON DELETE SET NULL;

-- 4. Drop old foreign key columns and the access_reviews table.
ALTER TABLE access_sources DROP COLUMN access_review_id;
ALTER TABLE access_review_campaigns DROP COLUMN access_review_id;

-- Update the LoadScopeSourcesByCampaignID query join: the old query joined
-- access_sources to access_review_campaigns via access_review_id. Now they
-- share organization_id directly.

DROP TABLE access_reviews;

-- 5. Move identity_source_id from organizations to access_review_campaigns.
ALTER TABLE access_review_campaigns ADD COLUMN identity_source_id TEXT REFERENCES access_sources(id);
ALTER TABLE organizations DROP COLUMN identity_source_id;
