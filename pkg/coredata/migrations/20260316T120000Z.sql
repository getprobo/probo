-- Add implemented state and not_implemented_justification columns to controls table
CREATE TYPE control_implementation_state AS ENUM ('IMPLEMENTED', 'NOT_IMPLEMENTED');
ALTER TABLE controls ADD COLUMN implemented control_implementation_state NOT NULL DEFAULT 'IMPLEMENTED';
ALTER TABLE controls ADD COLUMN not_implemented_justification text;
ALTER TABLE controls ALTER COLUMN implemented DROP DEFAULT;
