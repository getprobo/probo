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

ALTER TABLE iam_membership_profiles
    ADD COLUMN preferred_language TEXT,
    ADD COLUMN given_name TEXT,
    ADD COLUMN family_name TEXT,
    ADD COLUMN formatted_name TEXT,
    ADD COLUMN middle_name TEXT,
    ADD COLUMN honorific_prefix TEXT,
    ADD COLUMN honorific_suffix TEXT,
    ADD COLUMN employee_number TEXT,
    ADD COLUMN department TEXT,
    ADD COLUMN cost_center TEXT,
    ADD COLUMN enterprise_organization TEXT,
    ADD COLUMN division TEXT,
    ADD COLUMN manager_value TEXT;
