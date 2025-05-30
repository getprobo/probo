-- Rename policies table to documents
ALTER TABLE policies RENAME TO documents;

-- Rename risks_policies table to risks_documents
ALTER TABLE risks_policies RENAME TO risks_documents;

-- Update the foreign key reference in risks_documents
ALTER TABLE risks_documents RENAME CONSTRAINT risks_policies_policy_id_fkey TO risks_documents_document_id_fkey;

-- Rename the policy_id column in risks_documents to document_id
ALTER TABLE risks_documents RENAME COLUMN policy_id TO document_id;

-- Rename controls_policies table to controls_documents
ALTER TABLE controls_policies RENAME TO controls_documents;

-- Update controls_documents foreign key and column
ALTER TABLE controls_documents RENAME COLUMN policy_id TO document_id;
ALTER TABLE controls_documents RENAME CONSTRAINT controls_policies_policy_id_fkey TO controls_documents_document_id_fkey;

-- Rename policy_versions table to document_versions
ALTER TABLE policy_versions RENAME TO document_versions;

-- Update document_versions foreign key and column
ALTER TABLE document_versions RENAME COLUMN policy_id TO document_id;
ALTER TABLE document_versions RENAME CONSTRAINT policy_versions_policy_id_fkey TO document_versions_document_id_fkey;

-- Rename policy_version_signatures table to document_version_signatures
ALTER TABLE policy_version_signatures RENAME TO document_version_signatures;

-- Update document_version_signatures foreign key and column
ALTER TABLE document_version_signatures RENAME COLUMN policy_version_id TO document_version_id;
ALTER TABLE document_version_signatures RENAME CONSTRAINT policy_version_signatures_policy_version_id_fkey TO document_version_signatures_document_version_id_fkey;

-- Rename the unique index, preserving the WHERE clause
DROP INDEX policy_one_draft_version_idx;
CREATE UNIQUE INDEX document_one_draft_version_idx ON document_versions (document_id, status) WHERE status = 'DRAFT';
