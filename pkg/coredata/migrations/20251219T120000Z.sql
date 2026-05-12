-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

CREATE TYPE session_auth_method AS ENUM (
    'PASSWORD',
    'SAML'
);

ALTER TABLE sessions ADD COLUMN auth_method session_auth_method;
ALTER TABLE sessions ADD COLUMN authenticated_at TIMESTAMP;
ALTER TABLE sessions ADD COLUMN membership_id TEXT REFERENCES authz_memberships(id) ON DELETE CASCADE;

-- Expire all existing sessions as they are not backward compatible
UPDATE sessions SET 
    expire_reason = 'revoked',
    expired_at = NOW(),
    updated_at = NOW(),
    auth_method = 'PASSWORD',
    authenticated_at = created_at
WHERE expire_reason IS NULL;

UPDATE sessions SET 
    auth_method = 'PASSWORD',
    authenticated_at = created_at
WHERE auth_method IS NULL;

ALTER TABLE sessions ALTER COLUMN auth_method SET NOT NULL;
ALTER TABLE sessions ALTER COLUMN authenticated_at SET NOT NULL;
