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

-- Allow the prb CLI OAuth2 client to request the business-function scope.
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
    v1:business-function,
    v1:common-third-party,
    v1:compliance-page,
    v1:connector,
    v1:control,
    v1:datum,
    v1:document,
    v1:iam,
    v1:org,
    v1:privacy,
    v1:risk,
    v1:slack-connection,
    v1:task,
    v1:third-party,
    v1:webhook
}'::TEXT[],
    updated_at = NOW()
WHERE id = 'AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp';
