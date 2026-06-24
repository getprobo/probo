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
