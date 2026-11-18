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

-- A signature applies to a whole major version: minor publishes keep it and the
-- export unions signatures across every minor of the major. A signatory must
-- therefore have at most one signature per (document, major). Historically the
-- request guard was scoped to a single minor version, so re-requesting on a
-- newer minor (or twice on the same version) inserted duplicate rows, making a
-- person appear several times on the exported signature page.
--
-- Collapse each (document, major, signatory) group to a single signature,
-- keeping the same row the application now prefers: a SIGNED signature over a
-- pending one, then the most recently created.
WITH ranked AS (
    SELECT
        dvs.id,
        ROW_NUMBER() OVER (
            PARTITION BY dv.document_id, dv.major, dvs.signed_by_profile_id
            ORDER BY
                CASE dvs.state WHEN 'SIGNED' THEN 0 ELSE 1 END,
                dvs.created_at DESC,
                dvs.id DESC
        ) AS rn
    FROM document_version_signatures dvs
    INNER JOIN document_versions dv ON dvs.document_version_id = dv.id
)
DELETE FROM document_version_signatures
WHERE id IN (SELECT id FROM ranked WHERE rn > 1);
