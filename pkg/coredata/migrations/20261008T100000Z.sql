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

CREATE TABLE cookie_banner_discovery_stats (
    cookie_banner_id TEXT PRIMARY KEY REFERENCES cookie_banners(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    chrome_page_loads INTEGER NOT NULL DEFAULT 0,
    edge_page_loads INTEGER NOT NULL DEFAULT 0,
    firefox_page_loads INTEGER NOT NULL DEFAULT 0,
    safari_page_loads INTEGER NOT NULL DEFAULT 0,
    other_page_loads INTEGER NOT NULL DEFAULT 0,
    frozen_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE tracker_pattern_discovery_hits (
    tracker_pattern_id TEXT PRIMARY KEY REFERENCES tracker_patterns(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    chrome_hits INTEGER NOT NULL DEFAULT 0,
    edge_hits INTEGER NOT NULL DEFAULT 0,
    firefox_hits INTEGER NOT NULL DEFAULT 0,
    safari_hits INTEGER NOT NULL DEFAULT 0,
    other_hits INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
