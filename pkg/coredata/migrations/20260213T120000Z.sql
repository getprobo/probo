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

-- Drop the old columns
ALTER TABLE documents DROP COLUMN approver_profile_id;
ALTER TABLE document_versions DROP COLUMN approver_profile_id;
