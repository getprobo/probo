-- Create incident_registries table
CREATE TABLE incident_registries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    reference_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    source TEXT,
    owner_id TEXT NOT NULL REFERENCES peoples(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    incident_date TIMESTAMP WITH TIME ZONE,
    resolved_date TIMESTAMP WITH TIME ZONE,
    status TEXT NOT NULL CHECK (status IN ('OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED')) DEFAULT 'OPEN',
    priority TEXT NOT NULL CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')) DEFAULT 'MEDIUM',
    severity TEXT NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')) DEFAULT 'MEDIUM',
    category TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Ensure reference_id is unique per organization
    UNIQUE(organization_id, reference_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_incident_registries_organization_id ON incident_registries(organization_id);
CREATE INDEX idx_incident_registries_owner_id ON incident_registries(owner_id);
CREATE INDEX idx_incident_registries_status ON incident_registries(status);
CREATE INDEX idx_incident_registries_priority ON incident_registries(priority);
CREATE INDEX idx_incident_registries_severity ON incident_registries(severity);
CREATE INDEX idx_incident_registries_incident_date ON incident_registries(incident_date);
CREATE INDEX idx_incident_registries_created_at ON incident_registries(created_at);