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

-- Rename owner_profile_id to approver_profile_id on documents table
ALTER TABLE documents RENAME COLUMN owner_profile_id TO approver_profile_id;

-- Rename owner_profile_id to approver_profile_id on document_versions table
ALTER TABLE document_versions RENAME COLUMN owner_profile_id TO approver_profile_id;

-- Create document_approvers join table
CREATE TABLE document_approvers (
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    approver_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (document_id, approver_profile_id)
);

-- Create document_version_approvers join table
CREATE TABLE document_version_approvers (
    document_version_id TEXT NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
    approver_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (document_version_id, approver_profile_id)
);

-- Migrate existing approver data from documents to document_approvers
INSERT INTO document_approvers (document_id, approver_profile_id, tenant_id, organization_id, created_at)
SELECT id, approver_profile_id, tenant_id, organization_id, created_at
FROM documents
WHERE approver_profile_id IS NOT NULL
  AND deleted_at IS NULL;

-- Migrate existing approver data from document_versions to document_version_approvers
INSERT INTO document_version_approvers (document_version_id, approver_profile_id, tenant_id, organization_id, created_at)
SELECT id, approver_profile_id, tenant_id, organization_id, created_at
FROM document_versions
WHERE approver_profile_id IS NOT NULL;

-- Allow NULL on old columns now that data is in the join tables
ALTER TABLE document_versions ALTER COLUMN approver_profile_id DROP NOT NULL;
