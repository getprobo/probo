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
