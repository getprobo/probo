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

-- Enforce a single org ThirdParty per catalog vendor per organization.
--
-- The tracker-mapping worker used to materialize org third parties from
-- the catalog and could race the load-then-create check, leaving several
-- rows that share the same (organization_id, common_third_party_id).
-- Org third parties are now created only through the explicit import
-- action; this migration both cleans up the historical duplicates and
-- adds the partial unique index that makes the invariant enforceable.
--
-- DESTRUCTIVE: the DO block below merges duplicate org third parties onto
-- the earliest-created survivor by repointing every foreign key that
-- references third_parties(id) and then deleting the extras. It is driven
-- by catalog introspection (pg_constraint / pg_index) so it covers every
-- referencing table without hard-coding the list, and it removes link
-- rows that would collide on a referencing table's unique key before
-- repointing. Validate it against a production dump before deploying.

DO $$
DECLARE
    fk         RECORD;
    uq         RECORD;
    has_dupes  boolean;
BEGIN
    -- Map each duplicate row to the survivor it should be merged into:
    -- the earliest-created row for the (organization_id,
    -- common_third_party_id) pair, ties broken by id.
    CREATE TEMP TABLE _tp_dedupe ON COMMIT DROP AS
    SELECT t.id AS dup_id, k.survivor_id
    FROM third_parties t
    JOIN (
        SELECT
            organization_id,
            common_third_party_id,
            (array_agg(id ORDER BY created_at ASC, id ASC))[1] AS survivor_id
        FROM third_parties
        WHERE common_third_party_id IS NOT NULL
        GROUP BY organization_id, common_third_party_id
        HAVING count(*) > 1
    ) k
        ON t.organization_id = k.organization_id
       AND t.common_third_party_id = k.common_third_party_id
    WHERE t.id <> k.survivor_id;

    SELECT EXISTS (SELECT 1 FROM _tp_dedupe) INTO has_dupes;

    IF has_dupes THEN
        -- For every single-column foreign key that references
        -- third_parties(id) ...
        FOR fk IN
            SELECT c.conrelid::regclass::text AS tbl,
                   a.attname::text            AS col
            FROM pg_constraint c
            JOIN pg_attribute a
                ON a.attrelid = c.conrelid
               AND a.attnum = c.conkey[1]
            WHERE c.contype = 'f'
              AND c.confrelid = 'third_parties'::regclass
              AND array_length(c.conkey, 1) = 1
        LOOP
            -- ... drop duplicate-side rows that would collide with a
            -- survivor-side row on any unique key that includes the FK
            -- column (the survivor's row wins; this realizes the union of
            -- the two vendors' links).
            FOR uq IN
                SELECT (
                    SELECT string_agg(
                               format('s.%I IS NOT DISTINCT FROM d.%I',
                                      att.attname, att.attname),
                               ' AND ')
                    FROM pg_attribute att
                    WHERE att.attrelid = i.indrelid
                      AND att.attnum = ANY (i.indkey)
                      AND att.attname <> fk.col
                ) AS match_pred
                FROM pg_index i
                WHERE i.indrelid = fk.tbl::regclass
                  AND (i.indisunique OR i.indisprimary)
                  AND EXISTS (
                      SELECT 1
                      FROM pg_attribute att
                      WHERE att.attrelid = i.indrelid
                        AND att.attnum = ANY (i.indkey)
                        AND att.attname = fk.col
                  )
            LOOP
                IF uq.match_pred IS NOT NULL AND length(uq.match_pred) > 0 THEN
                    EXECUTE format(
                        'DELETE FROM %1$s d
                         USING _tp_dedupe m
                         WHERE d.%2$I = m.dup_id
                           AND EXISTS (
                               SELECT 1 FROM %1$s s
                               WHERE s.%2$I = m.survivor_id
                                 AND %3$s
                           )',
                        fk.tbl, fk.col, uq.match_pred
                    );
                END IF;
            END LOOP;

            -- Repoint the remaining references onto the survivor.
            EXECUTE format(
                'UPDATE %1$s d
                 SET %2$I = m.survivor_id
                 FROM _tp_dedupe m
                 WHERE d.%2$I = m.dup_id',
                fk.tbl, fk.col
            );
        END LOOP;

        -- Drop the now-unreferenced duplicate rows.
        DELETE FROM third_parties t
        USING _tp_dedupe m
        WHERE t.id = m.dup_id;
    END IF;
END $$;

CREATE UNIQUE INDEX third_parties_org_common_key
    ON third_parties (organization_id, common_third_party_id)
    WHERE common_third_party_id IS NOT NULL;
