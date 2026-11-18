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

CREATE TABLE common_gvl_snapshots (
    id                        TEXT PRIMARY KEY,
    vendor_list_version       INTEGER NOT NULL UNIQUE,
    gvl_specification_version INTEGER NOT NULL,
    tcf_policy_version        INTEGER NOT NULL,
    last_updated              TIMESTAMP WITH TIME ZONE NOT NULL,
    fetched_at                TIMESTAMP WITH TIME ZONE NOT NULL,
    payload                   JSONB NOT NULL
);

CREATE TABLE common_gvl_state (
    singleton                  BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    latest_vendor_list_version INTEGER REFERENCES common_gvl_snapshots (vendor_list_version),
    etag                       TEXT,
    cache_max_age_seconds      INTEGER,
    last_fetched_at            TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE common_gvl_vendors (
    id                    TEXT PRIMARY KEY,
    iab_vendor_id         INTEGER NOT NULL UNIQUE,
    vendor_list_version   INTEGER NOT NULL REFERENCES common_gvl_snapshots (vendor_list_version),
    name                  TEXT NOT NULL,
    deleted_date          TIMESTAMP WITH TIME ZONE,
    purposes              INTEGER[] NOT NULL DEFAULT '{}',
    leg_int_purposes      INTEGER[] NOT NULL DEFAULT '{}',
    flexible_purposes     INTEGER[] NOT NULL DEFAULT '{}',
    special_purposes      INTEGER[] NOT NULL DEFAULT '{}',
    features              INTEGER[] NOT NULL DEFAULT '{}',
    special_features      INTEGER[] NOT NULL DEFAULT '{}',
    policy_url            TEXT,
    uses_cookies          BOOLEAN,
    cookie_refresh        BOOLEAN,
    uses_non_cookie_access BOOLEAN,
    cookie_max_age_seconds INTEGER,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL
);
