-- Rename owner_profile_id to approver_profile_id on documents table
ALTER TABLE documents RENAME COLUMN owner_profile_id TO approver_profile_id;

-- Rename owner_profile_id to approver_profile_id on document_versions table
ALTER TABLE document_versions RENAME COLUMN owner_profile_id TO approver_profile_id;
