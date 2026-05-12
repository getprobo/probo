-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

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
