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

-- Dynamic registration must not create public clients. A public client skips
-- the organization membership check, so one organization could collect tokens
-- from users of another. Clients with no organization stay public: the Probo
-- CLI, and CIMD clients identified by external_client_id.
--
-- Refresh continues a grant and does not re-check membership. Revoke grants
-- issued while these clients were public, before the clients become private.
WITH downgraded AS (
    SELECT id
    FROM iam_oauth2_clients
    WHERE visibility = 'public'
        AND external_client_id IS NULL
        AND organization_id IS NOT NULL
),
revoked_refresh_tokens AS (
    UPDATE iam_oauth2_refresh_tokens
    SET revoked_at = NOW()
    WHERE revoked_at IS NULL
        AND client_id IN (SELECT id FROM downgraded)
    RETURNING id
),
deleted_access_tokens AS (
    DELETE FROM iam_oauth2_access_tokens
    WHERE client_id IN (SELECT id FROM downgraded)
    RETURNING id
),
deleted_authorization_codes AS (
    DELETE FROM iam_oauth2_authorization_codes
    WHERE client_id IN (SELECT id FROM downgraded)
    RETURNING id
),
deleted_device_codes AS (
    DELETE FROM iam_oauth2_device_codes
    WHERE client_id IN (SELECT id FROM downgraded)
    RETURNING id
)
UPDATE iam_oauth2_clients
SET visibility = 'private',
    updated_at = NOW()
WHERE id IN (SELECT id FROM downgraded);
