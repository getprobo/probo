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

ALTER TYPE obligations_status RENAME VALUE 'OPEN' TO 'NON_COMPLIANT';
ALTER TYPE obligations_status RENAME VALUE 'IN_PROGRESS' TO 'PARTIALLY_COMPLIANT';
ALTER TYPE obligations_status RENAME VALUE 'CLOSED' TO 'COMPLIANT';

ALTER TABLE obligations DROP COLUMN reference_id;

CREATE TABLE risks_obligations (
    risk_id TEXT NOT NULL REFERENCES risks(id) ON UPDATE CASCADE ON DELETE CASCADE,
    obligation_id TEXT NOT NULL REFERENCES obligations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (risk_id, obligation_id)
);

ALTER TABLE obligations ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    to_tsvector('simple',
        COALESCE(requirement, '') || ' ' ||
        COALESCE(area, '') || ' ' ||
        COALESCE(source, '') || ' ' ||
        COALESCE(regulator, '') || ' ' ||
        COALESCE(actions_to_be_implemented, '')
    )
) STORED;

CREATE INDEX obligations_search_idx ON obligations USING gin(search_vector);
