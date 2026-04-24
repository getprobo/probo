-- Add cookie_policy_url (required) and make privacy_policy_url optional.
-- Seed cookie_policy_url from existing privacy_policy_url for data continuity.

ALTER TABLE cookie_banners ADD COLUMN cookie_policy_url TEXT NOT NULL DEFAULT '';

UPDATE cookie_banners SET cookie_policy_url = privacy_policy_url WHERE cookie_policy_url = '';

ALTER TABLE cookie_banners ALTER COLUMN privacy_policy_url DROP NOT NULL;
