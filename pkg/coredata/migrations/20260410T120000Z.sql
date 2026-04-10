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

CREATE TYPE document_version_orientation AS ENUM ('PORTRAIT', 'LANDSCAPE');

CREATE TYPE document_content_source AS ENUM ('AUTHORED', 'GENERATED');

ALTER TABLE document_versions
    ADD COLUMN orientation document_version_orientation DEFAULT 'PORTRAIT';

ALTER TABLE document_versions
    ALTER COLUMN orientation DROP DEFAULT;

ALTER TABLE documents
    ADD COLUMN content_source document_content_source NOT NULL DEFAULT 'AUTHORED';

ALTER TABLE documents
    ALTER COLUMN content_source DROP DEFAULT;

ALTER TYPE document_type ADD VALUE IF NOT EXISTS 'STATEMENT_OF_APPLICABILITY';

ALTER TABLE statements_of_applicability
    ADD COLUMN document_id TEXT REFERENCES documents(id) ON DELETE SET NULL;

-- TODO: drop owner_profile_id column
ALTER TABLE statements_of_applicability
    ALTER COLUMN owner_profile_id DROP NOT NULL;

-- TODO: drop source_id column
-- TODO: drop snapshot_id column

CREATE TABLE statement_of_applicability_default_approvers (
    statement_of_applicability_id text                     NOT NULL,
    approver_profile_id           text                     NOT NULL,
    tenant_id                     text                     NOT NULL,
    organization_id               text                     NOT NULL,
    created_at                    timestamp with time zone NOT NULL,
    updated_at                    timestamp with time zone NOT NULL,
    PRIMARY KEY (statement_of_applicability_id, approver_profile_id),
    FOREIGN KEY (statement_of_applicability_id) REFERENCES statements_of_applicability(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (approver_profile_id) REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Backfill: create default approvers from existing owner_profile_id
INSERT INTO statement_of_applicability_default_approvers (
    statement_of_applicability_id, approver_profile_id, tenant_id, organization_id, created_at, updated_at
)
SELECT
    soa.id,
    soa.owner_profile_id,
    soa.tenant_id,
    soa.organization_id,
    NOW(),
    NOW()
FROM statements_of_applicability soa
WHERE soa.owner_profile_id IS NOT NULL
    AND soa.snapshot_id IS NULL;
