-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

-- Strip empty ProseMirror text nodes from document_versions.content.
-- Empty text nodes are invalid per the ProseMirror schema and cause Tiptap
-- to refuse rendering the document with "Empty text nodes are not allowed".

CREATE OR REPLACE FUNCTION pg_temp.strip_empty_text_nodes(input jsonb) RETURNS jsonb
LANGUAGE plpgsql IMMUTABLE AS $$
DECLARE
    item     jsonb;
    new_arr  jsonb := '[]'::jsonb;
    out_obj  jsonb;
BEGIN
    IF jsonb_typeof(input) = 'object' THEN
        out_obj := input;
        IF input ? 'content' AND jsonb_typeof(input->'content') = 'array' THEN
            FOR item IN SELECT * FROM jsonb_array_elements(input->'content') LOOP
                IF item->>'type' = 'text'
                   AND (item->>'text' IS NULL OR item->>'text' = '') THEN
                    CONTINUE;
                END IF;
                new_arr := new_arr || jsonb_build_array(pg_temp.strip_empty_text_nodes(item));
            END LOOP;
            out_obj := jsonb_set(out_obj, '{content}', new_arr);
        END IF;
        RETURN out_obj;
    ELSIF jsonb_typeof(input) = 'array' THEN
        FOR item IN SELECT * FROM jsonb_array_elements(input) LOOP
            new_arr := new_arr || jsonb_build_array(pg_temp.strip_empty_text_nodes(item));
        END LOOP;
        RETURN new_arr;
    ELSE
        RETURN input;
    END IF;
END;
$$;

DO $$
DECLARE
    r           record;
    cleaned     jsonb;
    cleaned_txt text;
    scanned     integer := 0;
    updated     integer := 0;
    skipped     integer := 0;
BEGIN
    FOR r IN
        SELECT dv.id, dv.content
        FROM document_versions dv
        JOIN documents d ON d.id = dv.document_id
        WHERE d.write_mode = 'GENERATED'
          AND (dv.content LIKE '%"text":""%' OR dv.content LIKE '%"text": ""%')
    LOOP
        scanned := scanned + 1;
        BEGIN
            cleaned := pg_temp.strip_empty_text_nodes(r.content::jsonb);
        EXCEPTION WHEN OTHERS THEN
            skipped := skipped + 1;
            RAISE NOTICE 'document_versions.id=% skipped: %', r.id, SQLERRM;
            CONTINUE;
        END;

        cleaned_txt := cleaned::text;
        IF cleaned_txt IS DISTINCT FROM r.content THEN
            UPDATE document_versions
            SET content = cleaned_txt
            WHERE id = r.id;
            updated := updated + 1;
        END IF;
    END LOOP;

    RAISE NOTICE 'strip_empty_text_nodes: scanned=% updated=% skipped=%',
        scanned, updated, skipped;
END $$;
