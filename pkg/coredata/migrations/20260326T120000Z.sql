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

ALTER TABLE document_versions DROP CONSTRAINT document_versions_document_id_version_number_key;

ALTER TABLE document_versions RENAME COLUMN version_number TO major;
ALTER TABLE document_versions ADD COLUMN minor INTEGER NOT NULL DEFAULT 0;
ALTER TABLE document_versions ALTER COLUMN minor DROP DEFAULT;

ALTER TABLE document_versions ADD CONSTRAINT document_versions_document_id_major_minor_key UNIQUE (document_id, major, minor);

ALTER TABLE documents RENAME COLUMN current_published_version TO current_published_major;
ALTER TABLE documents ADD COLUMN current_published_minor INTEGER;
UPDATE documents SET current_published_minor = 0 WHERE current_published_major IS NOT NULL;
