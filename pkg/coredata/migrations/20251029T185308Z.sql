-- Remove default_role column from auth_saml_configurations
-- The default role is now hardcoded to 'MEMBER' in the application code
-- when the SAML role attribute is missing or invalid

ALTER TABLE auth_saml_configurations DROP COLUMN default_role;
