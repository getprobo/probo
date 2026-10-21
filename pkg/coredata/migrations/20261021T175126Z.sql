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

-- The document row is the live document. Versions are frozen snapshots.
ALTER TABLE documents
    ADD COLUMN title TEXT,
    ADD COLUMN classification document_classification,
    ADD COLUMN document_type document_type,
    ADD COLUMN orientation document_version_orientation,
    ADD COLUMN content TEXT;

UPDATE documents AS d
SET
    title = lv.title,
    classification = lv.classification,
    document_type = lv.document_type,
    orientation = COALESCE(lv.orientation, 'PORTRAIT'),
    content = COALESCE(lv.content, '')
FROM (
    SELECT DISTINCT ON (document_id)
        document_id,
        title,
        classification,
        document_type,
        orientation,
        content
    FROM document_versions
    ORDER BY document_id, major DESC, minor DESC, created_at DESC
) AS lv
WHERE d.id = lv.document_id;

UPDATE documents
SET
    title = COALESCE(title, ''),
    classification = COALESCE(classification, 'CONFIDENTIAL'),
    document_type = COALESCE(document_type, 'OTHER'),
    orientation = COALESCE(orientation, 'PORTRAIT'),
    content = COALESCE(content, '')
WHERE
    title IS NULL
    OR classification IS NULL
    OR document_type IS NULL
    OR orientation IS NULL
    OR content IS NULL;

ALTER TABLE documents
    ALTER COLUMN title SET NOT NULL,
    ALTER COLUMN classification SET NOT NULL,
    ALTER COLUMN document_type SET NOT NULL,
    ALTER COLUMN orientation SET NOT NULL,
    ALTER COLUMN content SET NOT NULL;

ALTER TABLE documents
    ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('simple', COALESCE(title, ''))) STORED;

CREATE INDEX documents_search_idx ON documents USING gin(search_vector);

DELETE FROM document_versions WHERE status = 'DRAFT';

-- Postgres cannot drop an enum label. The partial index compares status, so
-- it has to go before the type is replaced.
DROP INDEX document_one_active_version_idx;

-- The same number can exist on a refused snapshot and the row that replaces it.
-- Only one published row may hold that number.
ALTER TABLE document_versions
    DROP CONSTRAINT document_versions_document_id_major_minor_key;

ALTER TYPE document_version_status RENAME TO document_version_status_old;

CREATE TYPE document_version_status AS ENUM (
    'PENDING_APPROVAL',
    'PUBLISHED',
    'REJECTED',
    'VOIDED'
);

ALTER TABLE document_versions
    ALTER COLUMN status TYPE document_version_status
    USING status::text::document_version_status;

DROP TYPE document_version_status_old;

CREATE UNIQUE INDEX document_versions_one_published_number_idx
    ON document_versions (document_id, major, minor)
    WHERE status = 'PUBLISHED';

CREATE UNIQUE INDEX document_one_pending_approval_idx
    ON document_versions (document_id)
    WHERE status = 'PENDING_APPROVAL';
