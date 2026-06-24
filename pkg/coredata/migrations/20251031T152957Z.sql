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

CREATE TYPE authz_api_role AS ENUM ('FULL');

CREATE TABLE auth_user_api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX auth_user_api_keys_user_id_idx ON auth_user_api_keys(user_id);

CREATE TABLE authz_api_keys_memberships (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    auth_user_api_key_id TEXT NOT NULL REFERENCES auth_user_api_keys(id) ON UPDATE CASCADE ON DELETE CASCADE,
    membership_id TEXT NOT NULL REFERENCES authz_memberships(id) ON UPDATE CASCADE ON DELETE CASCADE,
    role authz_api_role NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(auth_user_api_key_id, membership_id)
);

CREATE INDEX authz_api_keys_memberships_auth_user_api_key_id_idx ON authz_api_keys_memberships(auth_user_api_key_id);
CREATE INDEX authz_api_keys_memberships_membership_id_idx ON authz_api_keys_memberships(membership_id);
