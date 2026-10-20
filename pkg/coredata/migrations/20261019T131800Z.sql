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

CREATE TABLE cookie_banner_samples (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    chrome_page_loads INTEGER NOT NULL DEFAULT 0,
    edge_page_loads INTEGER NOT NULL DEFAULT 0,
    firefox_page_loads INTEGER NOT NULL DEFAULT 0,
    safari_page_loads INTEGER NOT NULL DEFAULT 0,
    other_page_loads INTEGER NOT NULL DEFAULT 0,
    chrome_consenting_page_loads INTEGER NOT NULL DEFAULT 0,
    edge_consenting_page_loads INTEGER NOT NULL DEFAULT 0,
    firefox_consenting_page_loads INTEGER NOT NULL DEFAULT 0,
    safari_consenting_page_loads INTEGER NOT NULL DEFAULT 0,
    other_consenting_page_loads INTEGER NOT NULL DEFAULT 0,
    frozen_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX cookie_banner_samples_open_uidx
    ON cookie_banner_samples (cookie_banner_id)
    WHERE frozen_at IS NULL;

CREATE INDEX cookie_banner_samples_banner_created_idx
    ON cookie_banner_samples (cookie_banner_id, created_at DESC);

CREATE TABLE tracker_pattern_sample_hits (
    sample_id TEXT NOT NULL REFERENCES cookie_banner_samples(id) ON DELETE CASCADE,
    tracker_pattern_id TEXT NOT NULL REFERENCES tracker_patterns(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    chrome_hits INTEGER NOT NULL DEFAULT 0,
    edge_hits INTEGER NOT NULL DEFAULT 0,
    firefox_hits INTEGER NOT NULL DEFAULT 0,
    safari_hits INTEGER NOT NULL DEFAULT 0,
    other_hits INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (sample_id, tracker_pattern_id)
);

-- One sample per legacy discovery window. Consenting page-load counters
-- did not exist on cookie_banner_discovery_stats, so they stay 0.
-- Hits can exist without a stats row, so those banners get their own
-- open sample. generate_gid entity type 138 = CookieBannerSampleEntityType.
WITH stats_inserted AS (
    INSERT INTO cookie_banner_samples (
        id,
        tenant_id,
        organization_id,
        cookie_banner_id,
        chrome_page_loads,
        edge_page_loads,
        firefox_page_loads,
        safari_page_loads,
        other_page_loads,
        chrome_consenting_page_loads,
        edge_consenting_page_loads,
        firefox_consenting_page_loads,
        safari_consenting_page_loads,
        other_consenting_page_loads,
        frozen_at,
        created_at,
        updated_at
    )
    SELECT
        generate_gid(decode_base64_unpadded(s.tenant_id), 138),
        s.tenant_id,
        b.organization_id,
        s.cookie_banner_id,
        s.chrome_page_loads,
        s.edge_page_loads,
        s.firefox_page_loads,
        s.safari_page_loads,
        s.other_page_loads,
        0,
        0,
        0,
        0,
        0,
        s.frozen_at,
        s.created_at,
        s.updated_at
    FROM cookie_banner_discovery_stats s
    JOIN cookie_banners b ON b.id = s.cookie_banner_id
    WHERE NOT EXISTS (
        SELECT 1
        FROM cookie_banner_samples cbs
        WHERE cbs.cookie_banner_id = s.cookie_banner_id
    )
    RETURNING id, tenant_id, cookie_banner_id
),
hit_only_inserted AS (
    INSERT INTO cookie_banner_samples (
        id,
        tenant_id,
        organization_id,
        cookie_banner_id,
        chrome_page_loads,
        edge_page_loads,
        firefox_page_loads,
        safari_page_loads,
        other_page_loads,
        chrome_consenting_page_loads,
        edge_consenting_page_loads,
        firefox_consenting_page_loads,
        safari_consenting_page_loads,
        other_consenting_page_loads,
        frozen_at,
        created_at,
        updated_at
    )
    SELECT
        generate_gid(decode_base64_unpadded(g.tenant_id), 138),
        g.tenant_id,
        b.organization_id,
        g.cookie_banner_id,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        NULL,
        g.created_at,
        g.updated_at
    FROM (
        SELECT
            h.cookie_banner_id,
            MIN(h.tenant_id) AS tenant_id,
            MIN(h.created_at) AS created_at,
            MAX(h.updated_at) AS updated_at
        FROM tracker_pattern_discovery_hits h
        WHERE NOT EXISTS (
            SELECT 1
            FROM cookie_banner_discovery_stats s
            WHERE s.cookie_banner_id = h.cookie_banner_id
        )
        AND NOT EXISTS (
            SELECT 1
            FROM cookie_banner_samples cbs
            WHERE cbs.cookie_banner_id = h.cookie_banner_id
        )
        GROUP BY h.cookie_banner_id
    ) g
    JOIN cookie_banners b ON b.id = g.cookie_banner_id
    RETURNING id, tenant_id, cookie_banner_id
),
inserted AS (
    SELECT id, tenant_id, cookie_banner_id FROM stats_inserted
    UNION ALL
    SELECT id, tenant_id, cookie_banner_id FROM hit_only_inserted
)
INSERT INTO tracker_pattern_sample_hits (
    sample_id,
    tracker_pattern_id,
    tenant_id,
    cookie_banner_id,
    chrome_hits,
    edge_hits,
    firefox_hits,
    safari_hits,
    other_hits,
    created_at,
    updated_at
)
SELECT
    i.id,
    h.tracker_pattern_id,
    h.tenant_id,
    h.cookie_banner_id,
    h.chrome_hits,
    h.edge_hits,
    h.firefox_hits,
    h.safari_hits,
    h.other_hits,
    h.created_at,
    h.updated_at
FROM tracker_pattern_discovery_hits h
JOIN inserted i
    ON i.cookie_banner_id = h.cookie_banner_id
    AND i.tenant_id = h.tenant_id;
