-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
