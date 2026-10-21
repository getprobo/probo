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

-- Prefer the draft version's Automerge state; otherwise the newest version that
-- has one. Documents with no state keep none: the CRDT is created on first open.
CREATE TEMP TABLE collaboration_backfill_source (
    document_id TEXT PRIMARY KEY,
    source_version_id TEXT NOT NULL
) ON COMMIT DROP;

INSERT INTO collaboration_backfill_source (
    document_id,
    source_version_id
)
SELECT DISTINCT ON (dv.document_id)
    dv.document_id,
    s.document_version_id
FROM document_versions dv
INNER JOIN document_version_automerge_states s
    ON s.document_version_id = dv.id
ORDER BY
    dv.document_id,
    CASE WHEN dv.status = 'DRAFT' THEN 0 ELSE 1 END,
    dv.major DESC,
    dv.minor DESC;

INSERT INTO collaboration_documents (
    tenant_id,
    document_id,
    organization_id,
    snapshot,
    heads,
    revision,
    snapshot_revision,
    change_revision,
    seeded,
    seed_claimed_at,
    created_at,
    updated_at
)
SELECT
    s.tenant_id,
    b.document_id,
    s.organization_id,
    s.snapshot,
    s.heads,
    s.revision,
    s.snapshot_revision,
    s.change_revision,
    s.seeded,
    s.seed_claimed_at,
    s.created_at,
    s.updated_at
FROM collaboration_backfill_source b
INNER JOIN document_version_automerge_states s
    ON s.document_version_id = b.source_version_id;

INSERT INTO collaboration_changes (
    tenant_id,
    document_id,
    organization_id,
    revision,
    change_hash,
    change_bytes,
    created_at
)
SELECT
    c.tenant_id,
    b.document_id,
    c.organization_id,
    c.revision,
    c.change_hash,
    c.change_bytes,
    c.created_at
FROM document_version_automerge_changes c
INNER JOIN collaboration_backfill_source b
    ON b.source_version_id = c.document_version_id;

INSERT INTO collaboration_actors (
    tenant_id,
    document_id,
    organization_id,
    actor_id,
    identity_id,
    profile_id,
    created_at
)
SELECT
    a.tenant_id,
    b.document_id,
    a.organization_id,
    a.actor_id,
    a.identity_id,
    a.profile_id,
    a.created_at
FROM document_version_automerge_actors a
INNER JOIN collaboration_backfill_source b
    ON b.source_version_id = a.document_version_id;

DROP TABLE document_version_collaboration_presences;
DROP TABLE document_version_automerge_actors;
DROP TABLE document_version_automerge_changes;
DROP TABLE document_version_automerge_states;
