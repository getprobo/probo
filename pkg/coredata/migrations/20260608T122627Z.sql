-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

CREATE TABLE risk_assessment_boundaries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    risk_assessment_scope_id TEXT NOT NULL REFERENCES risk_assessment_scopes(id) ON DELETE CASCADE,
    parent_boundary_id TEXT REFERENCES risk_assessment_boundaries(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT risk_assessment_boundaries_unique_name UNIQUE (risk_assessment_scope_id, name)
);

ALTER TABLE risk_assessment_nodes
    ADD COLUMN boundary_id TEXT REFERENCES risk_assessment_boundaries(id) ON DELETE SET NULL;

-- Migrate legacy BOUNDARY-typed nodes into the first-class boundary model.
-- Each boundary gets a freshly generated GID (entity type 101). Node names are
-- already unique per scope, so the boundary (scope_id, name) constraint cannot
-- be violated.
INSERT INTO risk_assessment_boundaries (
    id,
    tenant_id,
    organization_id,
    risk_assessment_scope_id,
    parent_boundary_id,
    name,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(tenant_id), 101),
    tenant_id,
    organization_id,
    risk_assessment_scope_id,
    NULL,
    name,
    created_at,
    updated_at
FROM risk_assessment_nodes
WHERE node_type = 'BOUNDARY';

-- Remove the legacy node representation now that the boundaries exist.
DELETE FROM risk_assessment_nodes
WHERE node_type = 'BOUNDARY';

-- Drop the now-unused BOUNDARY value from the node_type enum. PostgreSQL cannot
-- remove a value from an enum in place, so the type is rebuilt without it.
ALTER TYPE risk_assessment_node_type RENAME TO risk_assessment_node_type_old;

CREATE TYPE risk_assessment_node_type AS ENUM ('ENTITY', 'ASSET', 'DATA');

ALTER TABLE risk_assessment_nodes
    ALTER COLUMN node_type TYPE risk_assessment_node_type
    USING node_type::text::risk_assessment_node_type;

DROP TYPE risk_assessment_node_type_old;
