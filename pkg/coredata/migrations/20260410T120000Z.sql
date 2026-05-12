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

