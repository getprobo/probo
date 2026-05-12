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

CREATE TABLE trust_center_document_accesses(
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    trust_center_access_id TEXT NOT NULL REFERENCES trust_center_accesses(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    document_id TEXT REFERENCES documents(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    report_id TEXT REFERENCES reports(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (trust_center_access_id, document_id),
    UNIQUE (trust_center_access_id, report_id),
    CHECK ((document_id IS NOT NULL) != (report_id IS NOT NULL))
);
