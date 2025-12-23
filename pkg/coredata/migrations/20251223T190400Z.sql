CREATE INDEX IF NOT EXISTS idx_iam_memberships_identity_tenant_org ON iam_memberships(identity_id, tenant_id, organization_id);

