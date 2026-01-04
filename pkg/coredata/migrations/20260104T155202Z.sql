-- Create SCIM events table for debugging/audit
CREATE TABLE iam_scim_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    scim_configuration_id TEXT NOT NULL REFERENCES iam_scim_configurations(id) ON DELETE CASCADE,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    request_body TEXT,
    response_body TEXT,
    status_code INTEGER NOT NULL,
    error_message TEXT,
    membership_id TEXT REFERENCES iam_memberships(id) ON DELETE SET NULL,
    ip_address INET NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_iam_scim_events_tenant_id ON iam_scim_events(tenant_id);
CREATE INDEX idx_iam_scim_events_organization_id ON iam_scim_events(organization_id);
CREATE INDEX idx_iam_scim_events_scim_configuration_id ON iam_scim_events(scim_configuration_id);
CREATE INDEX idx_iam_scim_events_created_at ON iam_scim_events(created_at DESC);

