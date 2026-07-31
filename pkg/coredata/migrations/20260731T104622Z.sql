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

CREATE TABLE resource_tags (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations (id) ON UPDATE CASCADE ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    color TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT resource_tags_organization_id_key_key UNIQUE (organization_id, key),
    CONSTRAINT resource_tags_key_slug_check CHECK (key ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT resource_tags_color_hex_check CHECK (
        color IS NULL
        OR color ~ '^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$'
    )
);

CREATE TABLE resource_tag_assignments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    tag_id TEXT NOT NULL REFERENCES resource_tags (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT resource_tag_assignments_resource_id_tag_id_key UNIQUE (resource_id, tag_id)
);

UPDATE iam_oauth2_clients
SET scopes = '{
    openid,
    profile,
    email,
    offline_access,
    v1:access-review,
    v1:agent,
    v1:asset,
    v1:audit,
    v1:common-third-party,
    v1:compliance-page,
    v1:connector,
    v1:control,
    v1:datum,
    v1:document,
    v1:iam,
    v1:itam,
    v1:org,
    v1:privacy,
    v1:resource-tag,
    v1:risk,
    v1:slack-connection,
    v1:task,
    v1:third-party,
    v1:webhook
}'::TEXT[],
    updated_at = NOW()
WHERE id = 'AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp';
