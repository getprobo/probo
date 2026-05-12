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

CREATE TYPE control_maturity_level AS ENUM (
    'NONE',
    'INITIAL',
    'MANAGED',
    'DEFINED',
    'QUANTITATIVELY_MANAGED',
    'OPTIMIZING'
);

ALTER TABLE controls ADD COLUMN maturity_level control_maturity_level;

UPDATE controls SET maturity_level = CASE
    WHEN implemented = 'NOT_IMPLEMENTED' THEN 'NONE'::control_maturity_level
    ELSE 'INITIAL'::control_maturity_level
END;

ALTER TABLE controls ALTER COLUMN maturity_level SET NOT NULL;
ALTER TABLE controls ALTER COLUMN implemented DROP NOT NULL;

-- TODO: drop column and type in a future migration
-- ALTER TABLE controls DROP COLUMN implemented;
-- DROP TYPE control_implementation_state;
