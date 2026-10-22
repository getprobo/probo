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

-- Collaboration state is one CRDT per document, created lazily the first time
-- the document is opened for editing. Keying on document_id keeps document
-- creation untouched and lets deletion cascade.
CREATE TABLE collaboration_documents (
    tenant_id TEXT NOT NULL,
    document_id TEXT PRIMARY KEY REFERENCES documents(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    snapshot BYTEA NOT NULL,
    heads BYTEA NOT NULL,
    revision BIGINT NOT NULL,
    snapshot_revision BIGINT NOT NULL,
    change_revision BIGINT NOT NULL,
    seeded BOOLEAN NOT NULL,
    seed_claimed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE collaboration_changes (
    tenant_id TEXT NOT NULL,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    revision BIGINT NOT NULL,
    change_hash BYTEA NOT NULL CHECK (octet_length(change_hash) = 32),
    change_bytes BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (document_id, change_hash),
    UNIQUE (document_id, revision)
);

CREATE TABLE collaboration_actors (
    tenant_id TEXT NOT NULL,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    actor_id BYTEA NOT NULL,
    identity_id TEXT NOT NULL REFERENCES identities(id),
    profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (document_id, actor_id)
);
