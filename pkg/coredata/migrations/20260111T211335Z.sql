ALTER TABLE
    iam_saml_configurations
ALTER COLUMN
    email_domain TYPE CITEXT;

ALTER TABLE
    peoples
ALTER COLUMN
    primary_email_address TYPE CITEXT;

ALTER TABLE
    peoples
ALTER COLUMN
    additional_email_addresses TYPE CITEXT [];
