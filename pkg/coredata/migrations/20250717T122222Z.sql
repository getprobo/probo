CREATE TYPE control_status AS ENUM ('INCLUDED', 'EXCLUDED');

ALTER TABLE controls ADD COLUMN status control_status NOT NULL DEFAULT 'INCLUDED';
ALTER TABLE controls ALTER COLUMN status DROP DEFAULT;
ALTER TABLE controls ADD COLUMN exclusion_justification TEXT;
