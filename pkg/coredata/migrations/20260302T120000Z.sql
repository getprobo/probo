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
