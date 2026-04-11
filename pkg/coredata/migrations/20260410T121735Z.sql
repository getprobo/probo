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
