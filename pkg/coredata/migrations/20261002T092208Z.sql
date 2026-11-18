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

-- Hiding a trust-center file is deleting it. Drop leftover NONE rows, then
-- recreate the shared enum without that label. Access rows for NONE files
-- go with the file (cp_document_accesses.compliance_portal_file_id CASCADE).

DELETE FROM cp_files
WHERE compliance_portal_visibility = 'NONE';

DELETE FROM cp_documents
WHERE visibility = 'NONE';

DELETE FROM cp_audits
WHERE visibility = 'NONE';

-- CHECK (visibility IN (...)) binds literals to the old enum, so ALTER TYPE
-- fails unless those constraints are dropped first.
ALTER TABLE cp_documents
    DROP CONSTRAINT cp_documents_visibility_check;

ALTER TABLE cp_audits
    DROP CONSTRAINT cp_audits_visibility_check;

CREATE TYPE compliance_portal_visibility_new AS ENUM (
    'RESTRICTED',
    'PUBLIC'
);

ALTER TABLE cp_files
    ALTER COLUMN compliance_portal_visibility TYPE compliance_portal_visibility_new
        USING compliance_portal_visibility::text::compliance_portal_visibility_new;

ALTER TABLE cp_documents
    ALTER COLUMN visibility TYPE compliance_portal_visibility_new
        USING visibility::text::compliance_portal_visibility_new;

ALTER TABLE cp_audits
    ALTER COLUMN visibility TYPE compliance_portal_visibility_new
        USING visibility::text::compliance_portal_visibility_new;

DROP TYPE compliance_portal_visibility;
ALTER TYPE compliance_portal_visibility_new RENAME TO compliance_portal_visibility;
