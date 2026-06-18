-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

-- Well-known OAuth2 client for Auditor Mode (prbaud).
-- Read scopes plus v1:control for measure mutations. Keep in sync with
-- Constants.OAuth2.scopes in the Auditor Mode macOS app.
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
    'AAAAAAAAAAAASwAAAAAAAAAAcHJiYXVk',
    NULL,
    NULL,
    'Auditor Mode',
    'public',
    '{}',
    '{
        openid,
        profile,
        email,
        offline_access,
        v1:control:read,
        v1:org:read,
        v1:control
    }'::TEXT[],
    '{urn:ietf:params:oauth:grant-type:device_code,refresh_token}',
    '{code}',
    'none',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE
SET
    scopes = EXCLUDED.scopes,
    updated_at = NOW();
