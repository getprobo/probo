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

-- Audit PDFs moved to files via report_file_id in 20260603T000000Z. The
-- reports table and leftover report_id foreign keys were never dropped.

ALTER TABLE cp_document_accesses
    DROP CONSTRAINT cp_document_accesses_check;

ALTER TABLE cp_document_accesses
    DROP COLUMN report_id;

ALTER TABLE cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_check CHECK (
        (document_id IS NOT NULL)::int
        + (report_file_id IS NOT NULL)::int
        + (compliance_portal_file_id IS NOT NULL)::int
        = 1
    );

ALTER TABLE audits
    DROP COLUMN report_id;

DROP TABLE reports;
