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

-- Replace cookie_banner_state enum: DRAFT/PUBLISHED/DISABLED → ACTIVE/INACTIVE
CREATE TYPE cookie_banner_state_new AS ENUM ('ACTIVE', 'INACTIVE');

ALTER TABLE cookie_banners
    ALTER COLUMN state TYPE cookie_banner_state_new
    USING CASE
        WHEN state::text = 'DISABLED' THEN 'INACTIVE'::cookie_banner_state_new
        ELSE 'ACTIVE'::cookie_banner_state_new
    END;

DROP TYPE cookie_banner_state;
ALTER TYPE cookie_banner_state_new RENAME TO cookie_banner_state;

-- Version state for cookie banner configuration snapshots
CREATE TYPE cookie_banner_version_state AS ENUM ('DRAFT', 'PUBLISHED');

CREATE TABLE cookie_banner_versions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cookie_banner_id TEXT NOT NULL REFERENCES cookie_banners(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    state cookie_banner_version_state NOT NULL,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT cookie_banner_versions_banner_version_key
        UNIQUE (cookie_banner_id, version)
);

-- Denormalize organization_id onto categories and consent records
ALTER TABLE cookie_categories
    ADD COLUMN organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE cookie_consent_records
    ADD COLUMN organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    ADD COLUMN cookie_banner_version_id TEXT NOT NULL REFERENCES cookie_banner_versions(id);
