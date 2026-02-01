ALTER TABLE iam_scim_bridges ADD COLUMN last_synced_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE iam_scim_bridges ADD COLUMN next_sync_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
ALTER TABLE iam_scim_bridges ADD COLUMN sync_error TEXT;

CREATE INDEX idx_iam_scim_bridges_next_sync ON iam_scim_bridges (next_sync_at) WHERE state = 'ACTIVE';
