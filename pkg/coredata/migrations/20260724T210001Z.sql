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

DROP INDEX IF EXISTS idx_trust_center_accesses_identity_id_organization_id;

ALTER TABLE trust_center_accesses
    ADD CONSTRAINT trust_center_accesses_identity_id_trust_center_id_key
    UNIQUE (identity_id, trust_center_id);

ALTER TABLE trust_center_files
    ADD COLUMN trust_center_id TEXT REFERENCES trust_centers(id) ON UPDATE CASCADE ON DELETE CASCADE;

-- Attach every existing file to its organization's oldest trust center. Files
-- predate multi-portal support, so the oldest portal is the one they were
-- uploaded through; an administrator can move them afterwards. Deleting the
-- rows instead would also cascade away their trust_center_document_accesses.
UPDATE trust_center_files tcf
SET trust_center_id = (
    SELECT tc.id
    FROM trust_centers tc
    WHERE tc.organization_id = tcf.organization_id
    ORDER BY tc.created_at ASC, tc.id ASC
    LIMIT 1
);

-- Files whose organization has no trust center at all cannot satisfy the
-- foreign key. Fail loudly rather than silently destroying them.
DO $$
DECLARE
    orphan_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO orphan_count
    FROM trust_center_files
    WHERE trust_center_id IS NULL;

    IF orphan_count > 0 THEN
        RAISE EXCEPTION
            'cannot backfill trust_center_files.trust_center_id: % file(s) belong to an organization with no trust center; assign them to a trust center before re-running this migration',
            orphan_count;
    END IF;
END
$$;

ALTER TABLE trust_center_files
    ALTER COLUMN trust_center_id SET NOT NULL;
