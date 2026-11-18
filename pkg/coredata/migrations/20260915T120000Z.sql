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

-- Store risk analysis descriptions as ProseMirror/Tiptap JSON, matching tasks.

CREATE OR REPLACE FUNCTION pg_temp.plain_text_to_pm_doc(src text) RETURNS jsonb
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    line text;
    blocks jsonb := '[]'::jsonb;
    para jsonb;
BEGIN
    src := replace(replace(src, E'\r\n', E'\n'), E'\r', E'\n');

    FOREACH line IN ARRAY string_to_array(src, E'\n')
    LOOP
        IF line = '' THEN
            para := jsonb_build_object('type', 'paragraph');
        ELSE
            para := jsonb_build_object(
                'type', 'paragraph',
                'content', jsonb_build_array(
                    jsonb_build_object('type', 'text', 'text', line)
                )
            );
        END IF;

        blocks := blocks || jsonb_build_array(para);
    END LOOP;

    -- string_to_array('', E'\n') is empty; FromPlainText("") still emits
    -- one paragraph so Tiptap's block+ document schema can load it.
    IF jsonb_array_length(blocks) = 0 THEN
        blocks := jsonb_build_array(jsonb_build_object('type', 'paragraph'));
    END IF;

    RETURN jsonb_build_object('type', 'doc', 'content', blocks);
END;
$$;

ALTER TABLE risk_analyses
    ALTER COLUMN description TYPE jsonb
    USING CASE
        WHEN description IS NULL OR btrim(description, E' \t\n\r\v\f') = '' THEN NULL
        ELSE pg_temp.plain_text_to_pm_doc(description)
    END;
