ALTER TABLE controls ALTER COLUMN status DROP NOT NULL;

-- TODO: drop columns and type once all code references are fully removed
-- ALTER TABLE controls DROP COLUMN status;
-- ALTER TABLE controls DROP COLUMN exclusion_justification;
-- DROP TYPE control_status;
