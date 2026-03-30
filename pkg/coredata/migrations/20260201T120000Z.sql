-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

ALTER TABLE iam_scim_bridges ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0;
ALTER TABLE iam_scim_bridges ADD COLUMN total_sync_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE iam_scim_bridges ADD COLUMN total_failure_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE iam_scim_bridges ALTER COLUMN consecutive_failures DROP DEFAULT;
ALTER TABLE iam_scim_bridges ALTER COLUMN total_sync_count DROP DEFAULT;
ALTER TABLE iam_scim_bridges ALTER COLUMN total_failure_count DROP DEFAULT;

DROP INDEX IF EXISTS idx_iam_scim_bridges_next_sync;
CREATE INDEX idx_iam_scim_bridges_next_sync ON iam_scim_bridges (next_sync_at) WHERE state IN ('ACTIVE', 'FAILED', 'SYNCING');
