-- Create data table
CREATE TABLE data (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    data_sensitivity data_sensitivity NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    owner_id TEXT NOT NULL REFERENCES peoples(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create junction table for many-to-many relationship with vendors
CREATE TABLE data_vendors (
    datum_id TEXT NOT NULL REFERENCES data(id) ON DELETE CASCADE,
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (datum_id, vendor_id)
);
