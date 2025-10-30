CREATE TABLE processing_activity_vendors (
    processing_activity_id TEXT REFERENCES processing_activities(id) ON DELETE CASCADE,
    vendor_id TEXT REFERENCES vendors(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    snapshot_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (processing_activity_id, vendor_id)
);
