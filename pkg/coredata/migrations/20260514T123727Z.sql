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

ALTER TABLE common_third_parties ADD COLUMN slug TEXT;

UPDATE common_third_parties
SET slug = lower(
    trim(BOTH '-' FROM
        regexp_replace(
            regexp_replace(
                regexp_replace(name, '[^a-zA-Z0-9 _-]', '', 'g'),
            '[ _]+', '-', 'g'),
        '-+', '-', 'g')
    )
);

ALTER TABLE common_third_parties ALTER COLUMN slug SET NOT NULL;

DROP INDEX common_third_parties_name_key;

CREATE UNIQUE INDEX common_third_parties_slug_key ON common_third_parties (slug);
