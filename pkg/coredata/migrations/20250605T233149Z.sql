ALTER TABLE controls DROP COLUMN version;

ALTER TABLE controls RENAME COLUMN reference_id TO section_title;

CREATE OR REPLACE FUNCTION section_title_sort_key(text) RETURNS text AS $$
DECLARE
    result text := '';
    matches text[];
    remainder text := $1;
    current text;
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
    'Converts numbers in strings to zero-padded format for natural sorting, handling decimal numbers correctly';
