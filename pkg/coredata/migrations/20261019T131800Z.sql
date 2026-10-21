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
    frozen_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX cookie_banner_samples_open_uidx
    ON cookie_banner_samples (cookie_banner_id)
    WHERE frozen_at IS NULL;

CREATE INDEX cookie_banner_samples_banner_created_idx
    ON cookie_banner_samples (cookie_banner_id, created_at DESC);

-- presentation is OPT_IN, OPT_OUT, NOTICE, or '' for legacy discovery.
-- consent_action is a CookieConsentAction, or '' when the report omitted it.
CREATE TABLE cookie_banner_sample_page_loads (
    sample_id TEXT NOT NULL REFERENCES cookie_banner_samples(id) ON DELETE CASCADE,
    browser_family TEXT NOT NULL,
    presentation TEXT NOT NULL,
    consent_action TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    count INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (sample_id, browser_family, presentation, consent_action)
);

CREATE TABLE tracker_pattern_sample_hits (
    sample_id TEXT NOT NULL REFERENCES cookie_banner_samples(id) ON DELETE CASCADE,
    tracker_pattern_id TEXT NOT NULL REFERENCES tracker_patterns(id) ON DELETE CASCADE,
    browser_family TEXT NOT NULL,
    presentation TEXT NOT NULL,
    consent_action TEXT NOT NULL,
    category_granted BOOLEAN NOT NULL,
    tenant_id TEXT NOT NULL,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    count INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (
        sample_id,
        tracker_pattern_id,
        browser_family,
        presentation,
        consent_action,
        category_granted
    )
);

-- One sample per legacy discovery window. Those windows predate consent
-- buckets, so their counts are stored with an empty presentation and
-- consent action and category_granted = true.
-- generate_gid entity type 138 = CookieBannerSampleEntityType.
WITH stats_inserted AS (
    INSERT INTO cookie_banner_samples (
        id,
        tenant_id,
        organization_id,
        cookie_banner_id,
        frozen_at,
        created_at,
        updated_at
    )
    SELECT
        generate_gid(decode_base64_unpadded(s.tenant_id), 138),
        s.tenant_id,
        b.organization_id,
        s.cookie_banner_id,
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
    RETURNING id, tenant_id, cookie_banner_id, created_at, updated_at
),
hit_only_inserted AS (
    INSERT INTO cookie_banner_samples (
        id,
        tenant_id,
        organization_id,
        cookie_banner_id,
        frozen_at,
        created_at,
        updated_at
    )
    SELECT
        generate_gid(decode_base64_unpadded(g.tenant_id), 138),
        g.tenant_id,
        b.organization_id,
        g.cookie_banner_id,
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
    RETURNING id, tenant_id, cookie_banner_id, created_at, updated_at
),
inserted AS (
    SELECT id, tenant_id, cookie_banner_id, created_at, updated_at FROM stats_inserted
    UNION ALL
    SELECT id, tenant_id, cookie_banner_id, created_at, updated_at FROM hit_only_inserted
),
page_loads AS (
    INSERT INTO cookie_banner_sample_page_loads (
        sample_id,
        browser_family,
        presentation,
        consent_action,
        tenant_id,
        count,
        created_at,
        updated_at
    )
    SELECT
        i.id,
        v.browser_family,
        '',
        '',
        i.tenant_id,
        v.count,
        i.created_at,
        i.updated_at
    FROM cookie_banner_discovery_stats s
    JOIN inserted i
        ON i.cookie_banner_id = s.cookie_banner_id
        AND i.tenant_id = s.tenant_id
    CROSS JOIN LATERAL (
        VALUES
            ('CHROME', s.chrome_page_loads),
            ('EDGE', s.edge_page_loads),
            ('FIREFOX', s.firefox_page_loads),
            ('SAFARI', s.safari_page_loads),
            ('OTHER', s.other_page_loads)
    ) AS v(browser_family, count)
    WHERE v.count > 0
    RETURNING sample_id
)
INSERT INTO tracker_pattern_sample_hits (
    sample_id,
    tracker_pattern_id,
    browser_family,
    presentation,
    consent_action,
    category_granted,
    tenant_id,
    cookie_banner_id,
    count,
    created_at,
    updated_at
)
SELECT
    i.id,
    h.tracker_pattern_id,
    v.browser_family,
    '',
    '',
    TRUE,
    h.tenant_id,
    h.cookie_banner_id,
    v.count,
    h.created_at,
    h.updated_at
FROM tracker_pattern_discovery_hits h
JOIN inserted i
    ON i.cookie_banner_id = h.cookie_banner_id
    AND i.tenant_id = h.tenant_id
CROSS JOIN LATERAL (
    VALUES
        ('CHROME', h.chrome_hits),
        ('EDGE', h.edge_hits),
        ('FIREFOX', h.firefox_hits),
        ('SAFARI', h.safari_hits),
        ('OTHER', h.other_hits)
) AS v(browser_family, count)
WHERE v.count > 0;
