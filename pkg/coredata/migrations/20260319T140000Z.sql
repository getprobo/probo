ALTER TABLE organization_contexts ADD COLUMN product TEXT;
ALTER TABLE organization_contexts ADD COLUMN architecture TEXT;
ALTER TABLE organization_contexts ADD COLUMN team TEXT;
ALTER TABLE organization_contexts ADD COLUMN processes TEXT;
ALTER TABLE organization_contexts ADD COLUMN customers TEXT;
ALTER TABLE organization_contexts DROP COLUMN summary;
