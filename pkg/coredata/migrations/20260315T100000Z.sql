-- Add FINDINGS value to snapshots_type enum
-- This must be in its own migration because PostgreSQL requires the new enum
-- value to be committed before it can be used in DML statements.
ALTER TYPE snapshots_type ADD VALUE IF NOT EXISTS 'FINDINGS';
