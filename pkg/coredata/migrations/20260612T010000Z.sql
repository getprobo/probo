-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
