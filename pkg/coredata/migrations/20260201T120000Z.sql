ALTER TABLE iam_scim_bridges ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0;
ALTER TABLE iam_scim_bridges ADD COLUMN total_sync_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE iam_scim_bridges ADD COLUMN total_failure_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE iam_scim_bridges ALTER COLUMN consecutive_failures DROP DEFAULT;
ALTER TABLE iam_scim_bridges ALTER COLUMN total_sync_count DROP DEFAULT;
ALTER TABLE iam_scim_bridges ALTER COLUMN total_failure_count DROP DEFAULT;

DROP INDEX IF EXISTS idx_iam_scim_bridges_next_sync;
CREATE INDEX idx_iam_scim_bridges_next_sync ON iam_scim_bridges (next_sync_at) WHERE state IN ('ACTIVE', 'FAILED', 'SYNCING');
