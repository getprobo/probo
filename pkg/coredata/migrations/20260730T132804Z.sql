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

-- trust_centers branding briefly lived as page_title, then title, then
-- entity_name. Inserts only populate entity_name; leftover NOT NULL
-- page_title/title columns make organization create fail with
-- SQLSTATE 23502. Ensure entity_name is present and drop the obsolete
-- columns for installs that got stuck between those renames.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'trust_centers'
          AND column_name = 'entity_name'
    ) THEN
        ALTER TABLE trust_centers
            ADD COLUMN entity_name TEXT NOT NULL DEFAULT '';

        UPDATE trust_centers tc
        SET entity_name = o.name
        FROM organizations o
        WHERE tc.organization_id = o.id
          AND tc.entity_name = '';

        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'trust_centers'
              AND column_name = 'page_title'
        ) THEN
            UPDATE trust_centers
            SET entity_name = page_title
            WHERE entity_name = ''
              AND page_title IS NOT NULL
              AND page_title <> '';
        END IF;

        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'trust_centers'
              AND column_name = 'title'
        ) THEN
            UPDATE trust_centers
            SET entity_name = title
            WHERE entity_name = ''
              AND title IS NOT NULL
              AND title <> '';
        END IF;

        ALTER TABLE trust_centers
            ALTER COLUMN entity_name DROP DEFAULT;
    END IF;
END $$;

ALTER TABLE trust_centers
    DROP COLUMN IF EXISTS page_title;

ALTER TABLE trust_centers
    DROP COLUMN IF EXISTS title;
