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

CREATE TABLE compliance_frameworks (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id) ON UPDATE CASCADE ON DELETE CASCADE,
    framework_id TEXT NOT NULL REFERENCES frameworks(id) ON UPDATE CASCADE ON DELETE CASCADE,
    rank INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (trust_center_id, framework_id)
);

ALTER TABLE compliance_frameworks
ADD CONSTRAINT compliance_frameworks_trust_center_id_rank_key
    UNIQUE (trust_center_id, rank)
    DEFERRABLE INITIALLY DEFERRED;

INSERT INTO compliance_frameworks (
    id,
    tenant_id,
    organization_id,
    trust_center_id,
    framework_id,
    rank,
    created_at,
    updated_at
)
WITH distinct_frameworks AS (
    SELECT DISTINCT
        a.tenant_id,
        a.organization_id,
        a.framework_id
    FROM audits a
    WHERE a.trust_center_visibility IN ('PRIVATE', 'PUBLIC')
    AND EXISTS (
        SELECT 1 FROM trust_centers tc
        WHERE tc.organization_id = a.organization_id
    )
)
SELECT
    generate_gid(decode_base64_unpadded(df.tenant_id), 62),
    df.tenant_id,
    df.organization_id,
    tc.id AS trust_center_id,
    df.framework_id,
    ROW_NUMBER() OVER (
        PARTITION BY df.organization_id
        ORDER BY f.created_at ASC, df.framework_id ASC
    ) AS rank,
    NOW(),
    NOW()
FROM distinct_frameworks df
JOIN trust_centers tc ON tc.organization_id = df.organization_id
JOIN frameworks f ON f.id = df.framework_id;
