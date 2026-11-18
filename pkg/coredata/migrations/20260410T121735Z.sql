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

CREATE TYPE cookie_banner_state AS ENUM ('DRAFT', 'PUBLISHED', 'DISABLED');
CREATE TYPE cookie_consent_mode AS ENUM ('OPT_IN', 'OPT_OUT');
CREATE TYPE cookie_consent_action AS ENUM ('ACCEPT_ALL', 'REJECT_ALL', 'CUSTOMIZE', 'GPC');

CREATE TABLE cookie_banners (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    origin TEXT NOT NULL,
    state cookie_banner_state NOT NULL,
    privacy_policy_url TEXT NOT NULL,
    consent_expiry_days INTEGER NOT NULL,
    consent_mode cookie_consent_mode NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE cookie_categories (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    required BOOLEAN NOT NULL,
    rank INTEGER NOT NULL,
    cookies JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_cookie_categories_one_required_per_banner
    ON cookie_categories (cookie_banner_id) WHERE required = TRUE;

CREATE TABLE cookie_consent_records (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    visitor_id TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    consent_data JSONB NOT NULL,
    action cookie_consent_action NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
