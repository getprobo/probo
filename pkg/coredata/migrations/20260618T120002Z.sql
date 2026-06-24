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

-- Register API scopes on the Probo CLI OAuth2 client so the device
-- authorization flow can request v1:* scopes under enforcement.
UPDATE iam_oauth2_clients
SET scopes = '{
    openid,
    profile,
    email,
    offline_access,
    v1:access-review,
    v1:access-review:read,
    v1:agent,
    v1:agent:read,
    v1:asset,
    v1:asset:read,
    v1:audit,
    v1:audit:read,
    v1:common-third-party,
    v1:common-third-party:read,
    v1:compliance-page,
    v1:compliance-page:read,
    v1:connector,
    v1:connector:read,
    v1:control,
    v1:control:read,
    v1:datum,
    v1:datum:read,
    v1:document,
    v1:document:read,
    v1:iam,
    v1:iam:read,
    v1:org,
    v1:org:read,
    v1:privacy,
    v1:privacy:read,
    v1:risk,
    v1:risk:read,
    v1:slack-connection,
    v1:slack-connection:read,
    v1:task,
    v1:task:read,
    v1:third-party,
    v1:third-party:read,
    v1:webhook,
    v1:webhook:read
}'::TEXT[],
    updated_at = NOW()
WHERE id = 'AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp';
