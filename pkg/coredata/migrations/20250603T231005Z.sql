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

ALTER INDEX policy_versions_pkey RENAME TO document_versions_pkey;

ALTER TABLE document_versions
    RENAME CONSTRAINT policy_versions_policy_id_version_number_key
    TO document_versions_document_id_version_number_key;

ALTER TABLE document_versions
    RENAME CONSTRAINT policy_versions_published_by_fkey
    TO document_versions_published_by_fkey;

ALTER TABLE document_versions
    ADD CONSTRAINT document_versions_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES peoples(id)
    ON DELETE RESTRICT
    ON UPDATE CASCADE;

ALTER TABLE document_versions
    ADD COLUMN title TEXT,
    ADD COLUMN owner_id TEXT;

UPDATE document_versions dv
SET title = d.title,
    owner_id = d.owner_id
FROM documents d
WHERE dv.document_id = d.id;

ALTER TABLE document_versions
    ALTER COLUMN title SET NOT NULL,
    ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE document_versions
    ADD CONSTRAINT document_versions_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES peoples(id)
    ON DELETE RESTRICT
    ON UPDATE CASCADE;
