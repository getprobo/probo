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
