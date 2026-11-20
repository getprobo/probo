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

-- Add flow column to electronic_signatures for clickwrap vs signature distinction.
-- Existing rows are backfilled as SIGNATURE, then the default is dropped.

CREATE TYPE electronic_signature_flow AS ENUM ('SIGNATURE', 'CLICKWRAP');

ALTER TABLE electronic_signatures
    ADD COLUMN flow electronic_signature_flow NOT NULL DEFAULT 'SIGNATURE';

ALTER TABLE electronic_signatures
    ALTER COLUMN flow DROP DEFAULT;

-- Drop the old full-table unique constraint (applies to all flows).
ALTER TABLE electronic_signatures
    DROP CONSTRAINT electronic_signatures_organization_id_signer_email_document_type_file_id_key;

-- Replace with a partial unique index scoped to SIGNATURE flow only.
-- CLICKWRAP rows share the same file_id across subjects, so the old constraint
-- would collide on the second acceptance of the same version.
CREATE UNIQUE INDEX electronic_signatures_signature_unique_idx
    ON electronic_signatures (organization_id, signer_email, document_type, file_id)
    WHERE flow = 'SIGNATURE';

-- Add statement-of-applicability to electronic_signature_document_type enum
-- (missed in earlier migration; already in Go constants).
ALTER TYPE electronic_signature_document_type ADD VALUE IF NOT EXISTS 'STATEMENT_OF_APPLICABILITY';

-- Add context_sha256 for seal v2: hash of JCS-canonicalized context JSON.
-- NULL until the sealing worker processes the row.
ALTER TABLE electronic_signatures
    ADD COLUMN context_sha256 TEXT;
