-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

-- Backfill: set every draft version that has a pending approval quorum to PENDING_APPROVAL
UPDATE document_versions dv
SET status = 'PENDING_APPROVAL'
WHERE dv.status = 'DRAFT'
  AND EXISTS (
    SELECT 1
    FROM document_version_approval_quorums q
    WHERE q.version_id = dv.id
      AND q.status = 'PENDING'
  );

-- Backfill: void pending decisions in rejected or voided quorums
UPDATE document_version_approval_decisions d
SET state = 'VOIDED'
WHERE d.state = 'PENDING'
  AND EXISTS (
    SELECT 1
    FROM document_version_approval_quorums q
    WHERE q.id = d.quorum_id
      AND q.status IN ('REJECTED', 'VOIDED')
  );
