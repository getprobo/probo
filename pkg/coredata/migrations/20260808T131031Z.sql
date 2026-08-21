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

ALTER TABLE document_version_automerge_states
    ADD COLUMN snapshot_revision BIGINT;

ALTER TABLE document_version_automerge_states
    ADD COLUMN change_revision BIGINT;

ALTER TABLE document_version_automerge_states
    ADD COLUMN heads BYTEA;

UPDATE document_version_automerge_states
SET snapshot_revision = revision;

UPDATE document_version_automerge_states
SET change_revision = revision;

UPDATE document_version_automerge_states
SET heads = ''::BYTEA;

ALTER TABLE document_version_automerge_states
    ALTER COLUMN snapshot_revision SET NOT NULL;

ALTER TABLE document_version_automerge_states
    ALTER COLUMN change_revision SET NOT NULL;

ALTER TABLE document_version_automerge_states
    ALTER COLUMN heads SET NOT NULL;

CREATE TABLE document_version_automerge_changes (
    tenant_id TEXT NOT NULL,
    document_version_id TEXT NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    revision BIGINT NOT NULL,
    change_hash BYTEA NOT NULL CHECK (octet_length(change_hash) = 32),
    change_bytes BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (document_version_id, change_hash),
    UNIQUE (document_version_id, revision)
);
