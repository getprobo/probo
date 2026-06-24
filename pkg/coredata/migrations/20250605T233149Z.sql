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

ALTER TABLE controls DROP COLUMN version;

ALTER TABLE controls RENAME COLUMN reference_id TO section_title;

CREATE OR REPLACE FUNCTION section_title_sort_key(text) RETURNS text AS $$
DECLARE
    result text := '';
    matches text[];
    remainder text := $1;
BEGIN
    WHILE remainder ~ '\d+' LOOP
        -- Extract text before the number
        result := result || substring(remainder FROM '^[^\d]*');

        -- Extract and pad the next number
        matches := regexp_matches(remainder, '(\d+)', '');  -- captures the first number
        IF matches IS NOT NULL THEN
            result := result || lpad(matches[1], 10, '0');
            -- Remove processed part from remainder
            remainder := substring(remainder FROM '\d+(.*)$');
        ELSE
            EXIT;
        END IF;
    END LOOP;

    -- Append any remaining non-digit text
    result := result || remainder;

    RETURN result;
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;

COMMENT ON FUNCTION section_title_sort_key(text) IS
    'Converts numbers in strings to zero-padded format for natural sorting';
