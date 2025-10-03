CREATE TYPE trust_center_visibility AS ENUM (
    'NONE',
    'PRIVATE',
    'PUBLIC'
);

ALTER TABLE documents ADD COLUMN trust_center_visibility trust_center_visibility NOT NULL DEFAULT 'NONE';

UPDATE documents SET trust_center_visibility = CASE
    WHEN show_on_trust_center = true THEN 'PRIVATE'::trust_center_visibility
    ELSE 'NONE'::trust_center_visibility
END;

ALTER TABLE documents ALTER COLUMN trust_center_visibility DROP DEFAULT;
ALTER TABLE documents DROP COLUMN show_on_trust_center;

ALTER TABLE audits ADD COLUMN trust_center_visibility trust_center_visibility NOT NULL DEFAULT 'NONE';

UPDATE audits SET trust_center_visibility = CASE
    WHEN show_on_trust_center = true THEN 'PRIVATE'::trust_center_visibility
    ELSE 'NONE'::trust_center_visibility
END;

ALTER TABLE audits ALTER COLUMN trust_center_visibility DROP DEFAULT;
ALTER TABLE audits DROP COLUMN show_on_trust_center;
