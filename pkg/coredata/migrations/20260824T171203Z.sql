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

-- Give every connector a single owning feature. A connector credential is
-- used either by one access-review source or by one SCIM bridge, never
-- both, so each owner can delete its connector directly instead of
-- reference-counting under a row lock. Legacy rows that are shared
-- (several sources on one connector, or a source and a bridge on one
-- connector) are split by copying the connector row: the encrypted
-- connection is not bound to the row identity (AES-GCM with no additional
-- authenticated data), so a verbatim copy stays decryptable.

-- Split A: several sources on one connector. Every source ranked past the
-- oldest gets a copy of the connector. Sources are never deleted here —
-- campaign snapshot tables cascade off them.
CREATE TEMP TABLE tmp_source_connector_splits ON COMMIT DROP AS
SELECT
    ranked.id AS source_id,
    ranked.connector_id,
    generate_gid(decode_base64_unpadded(c.tenant_id), 5) AS new_connector_id
FROM (
    SELECT
        id,
        connector_id,
        ROW_NUMBER() OVER (
            PARTITION BY connector_id
            ORDER BY created_at ASC, id ASC
        ) AS rn
    FROM access_review_sources
    WHERE connector_id IS NOT NULL
) ranked
JOIN connectors c ON c.id = ranked.connector_id
WHERE ranked.rn > 1;

INSERT INTO connectors (
    id,
    tenant_id,
    organization_id,
    provider,
    protocol,
    settings,
    encrypted_connection,
    created_at,
    updated_at
)
SELECT
    t.new_connector_id,
    c.tenant_id,
    c.organization_id,
    c.provider,
    c.protocol,
    c.settings,
    c.encrypted_connection,
    c.created_at,
    NOW()
FROM tmp_source_connector_splits t
JOIN connectors c ON c.id = t.connector_id;

UPDATE access_review_sources s
SET connector_id = t.new_connector_id
FROM tmp_source_connector_splits t
WHERE s.id = t.source_id;

-- Split B: a source and a SCIM bridge on one connector. The bridge gets
-- its own copy; the remaining source keeps the original.
CREATE TEMP TABLE tmp_bridge_connector_splits ON COMMIT DROP AS
SELECT
    b.id AS bridge_id,
    c.id AS connector_id,
    generate_gid(decode_base64_unpadded(c.tenant_id), 5) AS new_connector_id
FROM iam_scim_bridges b
JOIN connectors c ON c.id = b.connector_id
WHERE EXISTS (
    SELECT 1 FROM access_review_sources s WHERE s.connector_id = c.id
);

INSERT INTO connectors (
    id,
    tenant_id,
    organization_id,
    provider,
    protocol,
    settings,
    encrypted_connection,
    created_at,
    updated_at
)
SELECT
    t.new_connector_id,
    c.tenant_id,
    c.organization_id,
    c.provider,
    c.protocol,
    c.settings,
    c.encrypted_connection,
    c.created_at,
    NOW()
FROM tmp_bridge_connector_splits t
JOIN connectors c ON c.id = t.connector_id;

UPDATE iam_scim_bridges b
SET connector_id = t.new_connector_id
FROM tmp_bridge_connector_splits t
WHERE b.id = t.bridge_id;

-- At most one source and one bridge per connector, enforced by the
-- schema; cross-feature exclusivity is checked at bind time. The source
-- index is what makes lock-free idempotent source creation safe.
CREATE UNIQUE INDEX idx_access_review_sources_connector_id
    ON access_review_sources (connector_id)
    WHERE connector_id IS NOT NULL;

CREATE UNIQUE INDEX idx_iam_scim_bridges_connector_id
    ON iam_scim_bridges (connector_id)
    WHERE connector_id IS NOT NULL;

-- Deleting a connector out from under a live bridge silently disabled SCIM
-- sync (SET NULL); refuse it instead, and delete the bridge before its
-- connector.
ALTER TABLE iam_scim_bridges DROP CONSTRAINT iam_scim_bridges_connector_id_fkey;
ALTER TABLE iam_scim_bridges ADD CONSTRAINT iam_scim_bridges_connector_id_fkey
    FOREIGN KEY (connector_id) REFERENCES connectors(id);
