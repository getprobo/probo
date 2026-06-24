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

CREATE TABLE cookies (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    cookie_category_id TEXT NOT NULL REFERENCES cookie_categories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    duration TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_cookies_unique_name_per_banner
    ON cookies (cookie_banner_id, name);

ALTER TABLE cookie_categories DROP COLUMN cookies;
