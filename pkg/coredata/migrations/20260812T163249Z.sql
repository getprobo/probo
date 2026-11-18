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

CREATE TABLE probot_identity_bindings (
    id                 TEXT        PRIMARY KEY,
    provider           TEXT        NOT NULL,
    external_tenant_id TEXT        NOT NULL,
    external_user_id   TEXT        NOT NULL,
    identity_id        TEXT        NOT NULL REFERENCES identities (id) ON DELETE CASCADE,
    created_at         TIMESTAMPTZ NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL,
    UNIQUE (provider, external_tenant_id, external_user_id),
    UNIQUE (identity_id, provider, external_tenant_id)
);

CREATE TABLE probot_identity_binding_challenges (
    hashed_token             BYTEA       PRIMARY KEY,
    provider                 TEXT        NOT NULL,
    external_tenant_id       TEXT        NOT NULL,
    external_user_id         TEXT        NOT NULL,
    expires_at               TIMESTAMPTZ NOT NULL,
    confirmed_by_identity_id TEXT        REFERENCES identities (id) ON DELETE CASCADE,
    confirmed_at             TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL
);

CREATE INDEX probot_identity_binding_challenges_expires_at_idx
    ON probot_identity_binding_challenges (expires_at);
