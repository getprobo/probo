-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
