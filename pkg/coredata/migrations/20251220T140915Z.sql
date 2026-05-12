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

CREATE TABLE iam_identity_profiles (
    tenant_id        TEXT NOT NULL,
    id               TEXT NOT NULL PRIMARY KEY,
    identity_id      TEXT NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    membership_id    TEXT REFERENCES iam_memberships(id) ON DELETE CASCADE,
    full_name        TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX idx_iam_identity_profiles_default ON iam_identity_profiles(identity_id) WHERE membership_id IS NULL;
CREATE UNIQUE INDEX idx_iam_identity_profiles_membership ON iam_identity_profiles(membership_id) WHERE membership_id IS NOT NULL;
