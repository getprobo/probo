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

ALTER TABLE trust_center_accesses ADD COLUMN has_accepted_non_disclosure_agreement BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE trust_center_accesses ALTER COLUMN has_accepted_non_disclosure_agreement DROP DEFAULT;
ALTER TABLE trust_center_accesses ADD COLUMN has_accepted_non_disclosure_agreement_metadata JSONB;

ALTER TABLE trust_centers ADD COLUMN non_disclosure_agreement_file_id TEXT;
ALTER TABLE trust_centers ADD CONSTRAINT trust_centers_non_disclosure_agreement_file_id_fkey
    FOREIGN KEY (non_disclosure_agreement_file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
