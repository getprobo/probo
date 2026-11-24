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

-- Every suggestion owns a comment thread from creation. Suggestions started
-- before that only had one once someone commented. The thread keeps no
-- anchor: the suggestion locates itself.

-- An open suggestion whose discussion was resolved gets it back.
UPDATE
    document_comment_threads
SET
    resolved_at = NULL,
    resolved_by_profile_id = NULL,
    updated_at = NOW()
WHERE
    id IN (
        SELECT DISTINCT ON (t.suggestion_id)
            t.id
        FROM
            document_comment_threads t
            INNER JOIN document_suggestions s ON s.id = t.suggestion_id
        WHERE
            s.status = 'OPEN'
            AND NOT EXISTS (
                SELECT
                    1
                FROM
                    document_comment_threads o
                WHERE
                    o.suggestion_id = t.suggestion_id
                    AND o.resolved_at IS NULL
            )
        ORDER BY
            t.suggestion_id,
            t.created_at DESC,
            t.id DESC
    );

-- An open suggestion with no thread gets an open one.
INSERT INTO document_comment_threads (
    tenant_id,
    id,
    organization_id,
    document_id,
    suggestion_id,
    anchor_start,
    anchor_end,
    anchor_quote,
    resolved_at,
    resolved_by_profile_id,
    created_at,
    updated_at
)
SELECT
    s.tenant_id,
    generate_gid(decode_base64_unpadded(s.tenant_id), 144),
    s.organization_id,
    s.document_id,
    s.id,
    NULL,
    NULL,
    '',
    NULL,
    NULL,
    s.created_at,
    s.created_at
FROM
    document_suggestions s
WHERE
    s.status = 'OPEN'
    AND NOT EXISTS (
        SELECT
            1
        FROM
            document_comment_threads t
        WHERE
            t.suggestion_id = s.id
    );

-- A decided suggestion with no thread gets one resolved by its decision.
INSERT INTO document_comment_threads (
    tenant_id,
    id,
    organization_id,
    document_id,
    suggestion_id,
    anchor_start,
    anchor_end,
    anchor_quote,
    resolved_at,
    resolved_by_profile_id,
    created_at,
    updated_at
)
SELECT
    s.tenant_id,
    generate_gid(decode_base64_unpadded(s.tenant_id), 144),
    s.organization_id,
    s.document_id,
    s.id,
    NULL,
    NULL,
    '',
    COALESCE(s.resolved_at, s.updated_at),
    s.resolved_by_profile_id,
    s.created_at,
    COALESCE(s.resolved_at, s.updated_at)
FROM
    document_suggestions s
WHERE
    s.status <> 'OPEN'
    AND NOT EXISTS (
        SELECT
            1
        FROM
            document_comment_threads t
        WHERE
            t.suggestion_id = s.id
    );

-- Threads that copied the suggestion's anchor drop it.
UPDATE
    document_comment_threads
SET
    anchor_start = NULL,
    anchor_end = NULL,
    anchor_quote = '',
    updated_at = NOW()
WHERE
    suggestion_id IS NOT NULL
    AND (
        anchor_start IS NOT NULL
        OR anchor_end IS NOT NULL
        OR anchor_quote <> ''
    );
