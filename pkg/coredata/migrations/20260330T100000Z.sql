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

-- Convert flag from single TEXT to TEXT array
ALTER TABLE access_entries
    ADD COLUMN flags TEXT[] NOT NULL DEFAULT '{}';

-- Migrate existing data: copy non-NONE flag values into the array
UPDATE access_entries
SET flags = ARRAY[flag]
WHERE flag != 'NONE';

-- Convert flag_reason to flag_reasons array
ALTER TABLE access_entries
    ADD COLUMN flag_reasons TEXT[] NOT NULL DEFAULT '{}';

-- Migrate existing flag_reason
UPDATE access_entries
SET flag_reasons = ARRAY[flag_reason]
WHERE flag_reason IS NOT NULL AND flag_reason != '';

-- Drop old columns in a single statement
ALTER TABLE access_entries
    DROP COLUMN flag,
    DROP COLUMN flag_reason;
