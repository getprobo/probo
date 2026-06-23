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

-- Narrow prb CLI OAuth2 client registration to unprefixed v1:* scopes.
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
