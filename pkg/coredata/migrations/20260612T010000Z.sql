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

-- Normalize access-review table, column, and enum names.

-- 1. Enum types
ALTER TYPE access_source_category RENAME TO access_review_source_category;
ALTER TYPE access_entry_decision RENAME TO access_review_entry_decision;
ALTER TYPE access_entry_incremental_tag RENAME TO access_review_entry_incremental_tag;
ALTER TYPE access_entry_flag RENAME TO access_review_entry_flag;

-- 2. Live sources
ALTER TABLE access_sources RENAME TO access_review_sources;

ALTER TABLE access_review_sources
    RENAME CONSTRAINT access_sources_pkey TO access_review_sources_pkey;
ALTER TABLE access_review_sources
    RENAME CONSTRAINT access_sources_organization_id_fkey TO access_review_sources_organization_id_fkey;
ALTER TABLE access_review_sources
    RENAME CONSTRAINT access_sources_connector_id_fkey TO access_review_sources_connector_id_fkey;

-- 3. Campaign source snapshots: live-source FK column
ALTER TABLE access_review_campaign_sources
    RENAME COLUMN access_source_id TO access_review_source_id;

ALTER TABLE access_review_campaign_sources
    RENAME CONSTRAINT access_review_campaign_sources_access_source_id_fkey
    TO access_review_campaign_sources_access_review_source_id_fkey;

ALTER TABLE access_review_campaign_sources
    DROP CONSTRAINT IF EXISTS access_review_campaign_sources_access_review_campaign_id_access_sour_key;
ALTER TABLE access_review_campaign_sources
    DROP CONSTRAINT IF EXISTS access_review_campaign_sources_access_review_campaign_id_access_source_id_key;
ALTER TABLE access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_sources_campaign_source_unique
    UNIQUE (access_review_campaign_id, access_review_source_id);

-- 4. Entries
ALTER TABLE access_entries RENAME TO access_review_entries;

ALTER TABLE access_review_entries
    RENAME CONSTRAINT access_entries_pkey TO access_review_entries_pkey;
ALTER TABLE access_review_entries
    RENAME CONSTRAINT access_entries_access_review_campaign_id_fkey
    TO access_review_entries_access_review_campaign_id_fkey;
ALTER TABLE access_review_entries
    RENAME CONSTRAINT access_entries_identity_id_fkey
    TO access_review_entries_identity_id_fkey;
ALTER TABLE access_review_entries
    RENAME CONSTRAINT access_entries_organization_id_fkey
    TO access_review_entries_organization_id_fkey;
ALTER TABLE access_review_entries
    RENAME CONSTRAINT access_entries_campaign_source_id_fkey
    TO access_review_entries_campaign_source_id_fkey;

ALTER INDEX idx_access_entries_campaign_source_account_key
    RENAME TO idx_access_review_entries_campaign_source_account_key;

-- 5. Decision history
ALTER TABLE access_entry_decision_history RENAME TO access_review_entry_decision_history;

ALTER TABLE access_review_entry_decision_history
    RENAME CONSTRAINT access_entry_decision_history_pkey
    TO access_review_entry_decision_history_pkey;
ALTER TABLE access_review_entry_decision_history
    RENAME CONSTRAINT access_entry_decision_history_access_entry_id_fkey
    TO access_review_entry_decision_history_access_review_entry_id_fkey;
ALTER TABLE access_review_entry_decision_history
    RENAME CONSTRAINT access_entry_decision_history_organization_id_fkey
    TO access_review_entry_decision_history_organization_id_fkey;

ALTER TABLE access_review_entry_decision_history
    RENAME COLUMN access_entry_id TO access_review_entry_id;
