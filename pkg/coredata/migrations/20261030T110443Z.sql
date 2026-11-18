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

-- A section title such as AC-2 is tokenized as a signed number, so a
-- search for the digit misses. Turn '-' and '+' into spaces before
-- to_tsvector. The list query does the same.

DROP INDEX controls_search_idx;

ALTER TABLE controls DROP COLUMN search_vector;

ALTER TABLE controls
    ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        to_tsvector(
            'simple',
            translate(
                COALESCE(section_title, '') || ' ' ||
                COALESCE(name, '') || ' ' ||
                COALESCE(description, ''),
                '-+', '  '
            )
        )
    ) STORED;

CREATE INDEX controls_search_idx ON controls USING gin(search_vector);
