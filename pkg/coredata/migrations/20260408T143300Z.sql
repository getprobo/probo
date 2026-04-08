-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
CREATE TYPE consent_action AS ENUM ('ACCEPT_ALL', 'REJECT_ALL', 'CUSTOMIZE', 'ACCEPT_CATEGORY', 'GPC');
CREATE TYPE consent_mode AS ENUM ('OPT_IN', 'OPT_OUT');

CREATE TABLE cookie_banners (
    id                     TEXT PRIMARY KEY,
    tenant_id              TEXT NOT NULL,
    organization_id        TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name                   TEXT NOT NULL,
    domain                 TEXT NOT NULL,
    state                  cookie_banner_state NOT NULL,
    title                  TEXT NOT NULL,
    description            TEXT NOT NULL,
    accept_all_label       TEXT NOT NULL,
    reject_all_label       TEXT NOT NULL,
    save_preferences_label TEXT NOT NULL,
    privacy_policy_url     TEXT NOT NULL,
    consent_expiry_days    INTEGER NOT NULL,
    version                INTEGER NOT NULL,
    theme                  JSONB,
    consent_mode           consent_mode NOT NULL,
    created_at             TIMESTAMPTZ NOT NULL,
    updated_at             TIMESTAMPTZ NOT NULL
);

CREATE TABLE cookie_categories (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name             TEXT NOT NULL,
    description      TEXT NOT NULL,
    required         BOOLEAN NOT NULL,
    rank             INTEGER NOT NULL,
    cookies          JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL
);

CREATE TABLE consent_records (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON UPDATE CASCADE ON DELETE CASCADE,
    visitor_id       TEXT NOT NULL,
    ip_address       TEXT,
    user_agent       TEXT,
    consent_data     JSONB NOT NULL,
    action           consent_action NOT NULL,
    banner_version   INTEGER NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL
);
