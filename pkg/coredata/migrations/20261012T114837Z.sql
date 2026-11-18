-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

-- Name resolution belongs to the connector account. The DEFAULT on
-- name_sync_attempts stays: probod deploys blue/green, and pods still running
-- the previous release insert accounts without this column until the rollout
-- finishes.
ALTER TABLE connector_accounts
    ADD COLUMN name_synced_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN name_sync_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN name_sync_next_attempt_at TIMESTAMP WITH TIME ZONE;

-- A synced source name is "<provider> / <instance>". The account keeps the
-- suffix. A name with no separator is left as stored. Sync state is copied
-- whenever the source was resolved or had spent attempts, so a given-up
-- account is not probed again. At most one source references an account.
UPDATE connector_accounts AS account
SET
    name = CASE
        WHEN source.name_synced_at IS NOT NULL
            AND strpos(source.name, ' / ') > 0
            AND substring(source.name FROM strpos(source.name, ' / ') + 3) <> ''
        THEN substring(source.name FROM strpos(source.name, ' / ') + 3)
        ELSE account.name
    END,
    name_synced_at = source.name_synced_at,
    name_sync_attempts = source.name_sync_attempts,
    name_sync_next_attempt_at = source.name_sync_next_attempt_at
FROM access_review_sources AS source
WHERE
    source.connector_account_id = account.id
    AND (
        source.name_synced_at IS NOT NULL
        OR source.name_sync_attempts > 0
    );

DROP INDEX idx_access_review_sources_name_sync;

-- The claim walks unsynced accounts in created_at order and stops at the
-- first row. This is the index that query had on access_review_sources.
CREATE INDEX idx_connector_accounts_name_sync
    ON connector_accounts (created_at)
    WHERE name_synced_at IS NULL;
