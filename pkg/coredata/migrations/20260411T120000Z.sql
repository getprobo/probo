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

-- Allow system-level OAuth2 clients that don't belong to any tenant or
-- organization (e.g. the Probo CLI).
ALTER TABLE iam_oauth2_clients ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE iam_oauth2_clients ALTER COLUMN organization_id DROP NOT NULL;

-- Well-known OAuth2 client for the Probo CLI (prb).
-- This client is hardcoded in the CLI binary and used for the device
-- authorization flow. Same pattern as GitHub CLI + GitHub Enterprise Server.
INSERT INTO iam_oauth2_clients (
    id,
    tenant_id,
    organization_id,
    client_name,
    visibility,
    redirect_uris,
    scopes,
    grant_types,
    response_types,
    token_endpoint_auth_method,
    created_at,
    updated_at
) VALUES (
    'AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp',
    NULL,
    NULL,
    'Probo CLI',
    'public',
    '{}',
    '{openid,profile,email}',
    '{urn:ietf:params:oauth:grant-type:device_code,refresh_token}',
    '{code}',
    'none',
    NOW(),
    NOW()
);
