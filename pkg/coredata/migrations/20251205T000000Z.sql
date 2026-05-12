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

CREATE TYPE trust_center_document_access_status AS ENUM ('REQUESTED', 'GRANTED', 'REJECTED', 'REVOKED');

ALTER TABLE
    trust_center_document_accesses
ADD
    COLUMN status trust_center_document_access_status;

DELETE FROM
    trust_center_document_accesses
WHERE
    active = 'F'
    AND requested = 'F';

UPDATE
    trust_center_document_accesses
SET
    status = CASE
        WHEN active = 'T' THEN 'GRANTED' :: trust_center_document_access_status
        ELSE 'REQUESTED' :: trust_center_document_access_status
    END;

ALTER TABLE
    trust_center_document_accesses
ALTER COLUMN
    status
SET
    NOT NULL;

ALTER TABLE
    trust_center_document_accesses DROP COLUMN active,
    DROP COLUMN requested;
