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

-- Rename control_state ENUM to mitigation_state
-- We need to create a new type and update all references since 
-- PostgreSQL doesn't support renaming enum types directly
CREATE TYPE mitigation_state AS ENUM (
    'NOT_STARTED',
    'IN_PROGRESS',
    'NOT_APPLICABLE',
    'IMPLEMENTED'
);

-- Create new mitigation_importance type
CREATE TYPE mitigation_importance AS ENUM (
    'MANDATORY', 
    'PREFERRED', 
    'ADVANCED'
);

-- Rename controls table to mitigations, adding the new columns with the new types
CREATE TABLE mitigations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    framework_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    content_ref TEXT NOT NULL,
    category TEXT NOT NULL,
    state mitigation_state NOT NULL,
    importance mitigation_importance NOT NULL,
    version INTEGER NOT NULL,
    standards TEXT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Copy data from controls to mitigations table
INSERT INTO mitigations (
    id, 
    tenant_id, 
    framework_id, 
    name, 
    description, 
    content_ref, 
    category, 
    state, 
    importance, 
    version,
    standards,
    created_at, 
    updated_at
)
SELECT 
    id, 
    tenant_id, 
    framework_id, 
    name, 
    description, 
    content_ref, 
    category, 
    (state::TEXT)::mitigation_state, 
    (importance::TEXT)::mitigation_importance, 
    version,
    standards,
    created_at, 
    updated_at
FROM controls;

-- Update tasks to reference mitigations instead of controls
ALTER TABLE tasks RENAME COLUMN control_id TO mitigation_id;

-- Update foreign key constraint
ALTER TABLE tasks DROP CONSTRAINT fk_tasks_control_id;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_mitigation_id
    FOREIGN KEY (mitigation_id) REFERENCES mitigations(id) ON DELETE CASCADE;

-- If we have any control_state_transitions table, rename it
-- Since the original migration file 20250310T161900Z.sql dropped it, this might not be necessary,
-- but including for completeness
DO $$
BEGIN
    IF EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_name = 'control_state_transitions'
    ) THEN
        ALTER TABLE control_state_transitions RENAME TO mitigation_state_transitions;
        ALTER TABLE mitigation_state_transitions RENAME COLUMN control_id TO mitigation_id;
        ALTER TABLE mitigation_state_transitions 
            DROP CONSTRAINT IF EXISTS control_state_transitions_control_id_fkey;
        ALTER TABLE mitigation_state_transitions 
            ADD CONSTRAINT mitigation_state_transitions_mitigation_id_fkey
            FOREIGN KEY (mitigation_id) REFERENCES mitigations(id);
    END IF;
END $$;

-- Drop old controls table and enum types after migration is complete
DROP TABLE controls;
DROP TYPE control_state;
DROP TYPE control_importance; 
