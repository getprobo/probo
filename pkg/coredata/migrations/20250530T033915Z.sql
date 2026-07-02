-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
