UPDATE peoples
SET additional_email_addresses = ARRAY[]::TEXT[]
WHERE additional_email_addresses IS NULL;

ALTER TABLE peoples ALTER COLUMN additional_email_addresses SET NOT NULL;
