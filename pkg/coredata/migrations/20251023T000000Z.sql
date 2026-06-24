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
