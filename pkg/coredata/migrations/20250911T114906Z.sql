ALTER TABLE framework_exports ADD COLUMN recipient_email CITEXT NOT NULL DEFAULT '';

ALTER TABLE framework_exports ALTER COLUMN recipient_email DROP DEFAULT;

ALTER TABLE framework_exports ADD COLUMN recipient_name TEXT NOT NULL DEFAULT '';

ALTER TABLE framework_exports ALTER COLUMN recipient_name DROP DEFAULT;
