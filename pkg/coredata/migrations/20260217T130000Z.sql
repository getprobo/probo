ALTER TABLE documents DROP COLUMN approver_profile_id;
ALTER TABLE document_versions DROP COLUMN approver_profile_id;

ALTER TABLE controls DROP COLUMN status;
ALTER TABLE controls DROP COLUMN exclusion_justification;
DROP TYPE control_status;
