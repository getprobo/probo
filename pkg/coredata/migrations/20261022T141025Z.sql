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

DELETE FROM document_versions WHERE status = 'DRAFT';

-- Postgres cannot drop an enum label. The partial index compares status, so
-- it has to go before the type is replaced.
DROP INDEX document_one_active_version_idx;

-- The same number can exist on a refused snapshot and the row that replaces it.
-- Only one published row may hold that number.
ALTER TABLE document_versions
    DROP CONSTRAINT document_versions_document_id_major_minor_key;

ALTER TYPE document_version_status RENAME TO document_version_status_old;

CREATE TYPE document_version_status AS ENUM (
    'PENDING_APPROVAL',
    'PUBLISHED',
    'REJECTED',
    'VOIDED'
);

ALTER TABLE document_versions
    ALTER COLUMN status TYPE document_version_status
    USING status::text::document_version_status;

DROP TYPE document_version_status_old;

CREATE UNIQUE INDEX document_versions_one_published_number_idx
    ON document_versions (document_id, major, minor)
    WHERE status = 'PUBLISHED';

CREATE UNIQUE INDEX document_one_pending_approval_idx
    ON document_versions (document_id)
    WHERE status = 'PENDING_APPROVAL';
