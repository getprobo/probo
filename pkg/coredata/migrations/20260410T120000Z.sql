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

CREATE TYPE document_version_orientation AS ENUM ('PORTRAIT', 'LANDSCAPE');

CREATE TYPE document_write_mode AS ENUM ('AUTHORED', 'GENERATED');

ALTER TABLE document_versions
    ADD COLUMN orientation document_version_orientation DEFAULT 'PORTRAIT';

ALTER TABLE document_versions
    ALTER COLUMN orientation DROP DEFAULT;

ALTER TABLE documents
    ADD COLUMN write_mode document_write_mode NOT NULL DEFAULT 'AUTHORED';

ALTER TABLE documents
    ALTER COLUMN write_mode DROP DEFAULT;

ALTER TYPE document_type ADD VALUE 'STATEMENT_OF_APPLICABILITY';

ALTER TYPE electronic_signature_document_type ADD VALUE 'STATEMENT_OF_APPLICABILITY';

ALTER TABLE statements_of_applicability
    ADD COLUMN document_id TEXT UNIQUE REFERENCES documents(id) ON DELETE SET NULL;

-- TODO: drop owner_profile_id column
ALTER TABLE statements_of_applicability
    ALTER COLUMN owner_profile_id DROP NOT NULL;

-- TODO: drop statements_of_applicability.source_id column
-- TODO: drop statements_of_applicability.snapshot_id column
-- TODO: drop applicability_statements.snapshot_id column

