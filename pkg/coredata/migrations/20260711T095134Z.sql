-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

-- Archived documents must not retain in-flight approval or signature workflows.
-- Backfill rows that were archived before archive teardown was enforced in the
-- application layer.

UPDATE document_version_approval_decisions d
SET
    state = 'VOIDED',
    updated_at = NOW()
FROM document_version_approval_quorums q
INNER JOIN document_versions dv ON dv.id = q.version_id
INNER JOIN documents doc ON doc.id = dv.document_id
WHERE
    d.quorum_id = q.id
    AND d.state = 'PENDING'
    AND doc.archived_at IS NOT NULL;

UPDATE document_version_approval_quorums q
SET
    status = 'VOIDED',
    updated_at = NOW()
FROM document_versions dv
INNER JOIN documents doc ON doc.id = dv.document_id
WHERE
    q.version_id = dv.id
    AND q.status = 'PENDING'
    AND doc.archived_at IS NOT NULL;

UPDATE document_versions dv
SET
    status = 'DRAFT',
    major = CASE
        WHEN doc.current_published_major IS NOT NULL THEN doc.current_published_major
        ELSE 0
    END,
    minor = CASE
        WHEN doc.current_published_major IS NOT NULL THEN doc.current_published_minor + 1
        ELSE 1
    END,
    updated_at = NOW()
FROM documents doc
WHERE
    dv.document_id = doc.id
    AND doc.archived_at IS NOT NULL
    AND dv.status = 'PENDING_APPROVAL';

DELETE FROM document_version_signatures dvs
USING document_versions dv
INNER JOIN documents doc ON doc.id = dv.document_id
WHERE
    dvs.document_version_id = dv.id
    AND dvs.state = 'REQUESTED'
    AND doc.archived_at IS NOT NULL;
