-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

CREATE TABLE trust_center_files (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    file_id TEXT NOT NULL REFERENCES files(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    trust_center_visibility trust_center_visibility NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

ALTER TABLE trust_center_document_accesses ADD COLUMN trust_center_file_id TEXT REFERENCES trust_center_files(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE trust_center_document_accesses DROP CONSTRAINT trust_center_document_accesses_check;

ALTER TABLE trust_center_document_accesses ADD CONSTRAINT trust_center_document_accesses_check CHECK (
    (document_id IS NOT NULL)::int + (report_id IS NOT NULL)::int + (trust_center_file_id IS NOT NULL)::int = 1
);

ALTER TABLE trust_center_document_accesses ADD CONSTRAINT trust_center_document_accesses_trust_center_file_id_key UNIQUE (trust_center_access_id, trust_center_file_id);
