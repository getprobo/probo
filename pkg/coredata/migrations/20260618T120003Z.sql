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
