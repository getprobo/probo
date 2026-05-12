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

ALTER TABLE trust_center_accesses ADD COLUMN has_accepted_non_disclosure_agreement BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE trust_center_accesses ALTER COLUMN has_accepted_non_disclosure_agreement DROP DEFAULT;
ALTER TABLE trust_center_accesses ADD COLUMN has_accepted_non_disclosure_agreement_metadata JSONB;

ALTER TABLE trust_centers ADD COLUMN non_disclosure_agreement_file_id TEXT;
ALTER TABLE trust_centers ADD CONSTRAINT trust_centers_non_disclosure_agreement_file_id_fkey
    FOREIGN KEY (non_disclosure_agreement_file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
