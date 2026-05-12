-- Copyright (c) 2026 Probo Inc <hello@probo.inc>
--
-- Permission to use, copy, modify, and distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
-- WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
-- MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
-- ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
-- WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
-- ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
-- OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

-- The index constraint is already respected in Probo's production environment.
-- The deletion is here in case duplicates were created directly via SQL in
-- other contexts.
DELETE FROM document_versions
WHERE status IN ('DRAFT', 'PENDING_APPROVAL')
  AND document_id IN (
    SELECT document_id
    FROM document_versions
    WHERE status IN ('DRAFT', 'PENDING_APPROVAL')
    GROUP BY document_id
    HAVING COUNT(*) > 1
  )
  AND id NOT IN (
    SELECT DISTINCT ON (document_id) id
    FROM document_versions
    WHERE status IN ('DRAFT', 'PENDING_APPROVAL')
    ORDER BY document_id, CASE WHEN status = 'PENDING_APPROVAL' THEN 0 ELSE 1 END, created_at ASC
  );

DROP INDEX document_one_draft_version_idx;

CREATE UNIQUE INDEX document_one_active_version_idx
    ON document_versions (document_id)
    WHERE status IN ('DRAFT', 'PENDING_APPROVAL');
