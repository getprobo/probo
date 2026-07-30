-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

-- title (or page_title, if the rename was skipped) is the full home page
-- heading. Rows that still store the org name as a fragment (the
-- pre-migration default) are backfilled to the English hero copy the UI
-- previously rendered via i18n ("Compliance at {{name}}.").
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'trust_centers'
          AND column_name = 'title'
    ) THEN
        UPDATE trust_centers tc
        SET title = 'Compliance at ' || o.name || '.',
            updated_at = NOW()
        FROM organizations o
        WHERE tc.organization_id = o.id
          AND tc.title = o.name;
    ELSIF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'trust_centers'
          AND column_name = 'page_title'
    ) THEN
        UPDATE trust_centers tc
        SET page_title = 'Compliance at ' || o.name || '.',
            updated_at = NOW()
        FROM organizations o
        WHERE tc.organization_id = o.id
          AND tc.page_title = o.name;
    END IF;
END $$;
