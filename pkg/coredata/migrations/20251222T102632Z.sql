-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

ALTER TABLE identities ADD COLUMN full_name TEXT NOT NULL DEFAULT '';

UPDATE identities i
SET full_name = COALESCE(
    (SELECT p.full_name FROM iam_identity_profiles p 
     WHERE p.identity_id = i.id AND p.membership_id IS NULL),
    ''
);

DELETE FROM iam_identity_profiles WHERE membership_id IS NULL;
DROP INDEX IF EXISTS idx_iam_identity_profiles_default;

ALTER TABLE iam_identity_profiles DROP COLUMN identity_id;

ALTER TABLE iam_identity_profiles RENAME TO iam_membership_profiles;
