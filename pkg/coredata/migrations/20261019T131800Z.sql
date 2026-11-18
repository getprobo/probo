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

-- presentation is OPT_IN, OPT_OUT, NOTICE, or '' when the report omitted it.
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
