DROP INDEX IF EXISTS idx_auth_saml_assertions_tenant_id;
ALTER TABLE iam_saml_assertions DROP COLUMN IF EXISTS tenant_id;