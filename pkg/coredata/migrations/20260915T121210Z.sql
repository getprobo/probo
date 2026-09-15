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

ALTER TABLE risks ADD COLUMN reference_id TEXT NOT NULL DEFAULT '';

UPDATE risks
SET reference_id = 'RSK-' || LPAD(seq.n::TEXT, 3, '0')
FROM (
	SELECT
		id,
		ROW_NUMBER() OVER (PARTITION BY organization_id ORDER BY created_at, id) AS n
	FROM risks
) seq
WHERE risks.id = seq.id;

ALTER TABLE risks ALTER COLUMN reference_id DROP DEFAULT;

ALTER TABLE risks
	ADD CONSTRAINT risks_organization_id_reference_id_key
	UNIQUE (organization_id, reference_id);

DROP INDEX IF EXISTS risks_search_idx;

ALTER TABLE risks DROP COLUMN search_vector;

ALTER TABLE risks ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
	to_tsvector('simple',
		COALESCE(reference_id, '') || ' ' || COALESCE(name, '')
	)
) STORED;

CREATE INDEX risks_search_idx ON risks USING gin(search_vector);
