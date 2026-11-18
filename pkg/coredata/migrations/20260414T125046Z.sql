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

-- Deduplicate connectors per (organization_id, provider), keeping the oldest one.
-- Temporarily switch access_sources FK to CASCADE for cleanup, then restore RESTRICT.

-- Swap FK to CASCADE so deleting duplicate connectors cascades to their access_sources.
ALTER TABLE access_sources DROP CONSTRAINT access_sources_connector_id_fkey;
ALTER TABLE access_sources ADD CONSTRAINT access_sources_connector_id_fkey
    FOREIGN KEY (connector_id) REFERENCES connectors(id) ON DELETE CASCADE;

-- Delete the duplicate connectors (access_sources will cascade).
DELETE FROM connectors
WHERE id IN (
    SELECT id FROM (
        SELECT id,
               ROW_NUMBER() OVER (
                   PARTITION BY organization_id, provider
                   ORDER BY created_at ASC
               ) AS rn
        FROM connectors
    ) ranked
    WHERE rn > 1
);

-- Restore FK to RESTRICT.
ALTER TABLE access_sources DROP CONSTRAINT access_sources_connector_id_fkey;
ALTER TABLE access_sources ADD CONSTRAINT access_sources_connector_id_fkey
    FOREIGN KEY (connector_id) REFERENCES connectors(id);

CREATE UNIQUE INDEX idx_connectors_organization_id_provider ON connectors (organization_id, provider);
