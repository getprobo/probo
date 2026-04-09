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
