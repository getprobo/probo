CREATE TABLE iam_scim_bridges (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    scim_configuration_id TEXT NOT NULL REFERENCES iam_scim_configurations(id) ON DELETE CASCADE,
    connector_id TEXT REFERENCES connectors(id) ON DELETE SET NULL,
    type TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT iam_scim_bridges_scim_configuration_unique UNIQUE (scim_configuration_id)
);
