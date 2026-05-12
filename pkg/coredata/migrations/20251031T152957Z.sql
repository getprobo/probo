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
