-- Add failure tracking columns to iam_scim_bridges table
ALTER TABLE iam_scim_bridges ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0;
ALTER TABLE iam_scim_bridges ADD COLUMN total_sync_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE iam_scim_bridges ADD COLUMN total_failure_count INTEGER NOT NULL DEFAULT 0;

-- Drop old index and create new one that covers all processable states
DROP INDEX IF EXISTS idx_iam_scim_bridges_next_sync;
CREATE INDEX idx_iam_scim_bridges_next_sync ON iam_scim_bridges (next_sync_at) WHERE state IN ('ACTIVE', 'FAILED', 'SYNCING');
