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
