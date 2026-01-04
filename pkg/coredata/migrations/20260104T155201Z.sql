-- Create SCIM configurations table
CREATE TABLE iam_scim_configurations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    hashed_token BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT iam_scim_configurations_organization_unique UNIQUE (organization_id)
);

CREATE INDEX idx_iam_scim_configurations_tenant_id ON iam_scim_configurations(tenant_id);
CREATE INDEX idx_iam_scim_configurations_organization_id ON iam_scim_configurations(organization_id);

