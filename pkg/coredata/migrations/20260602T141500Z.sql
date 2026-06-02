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

-- Drop dead schema left over from completed migrations. None of the objects
-- below are referenced by any query code, GraphQL resolver, or frontend; they
-- only appear in older migration files.

-- users_organizations was superseded by iam_memberships (data was copied over
-- in 20251006T220024Z). The legacy junction table is no longer read or written.
DROP TABLE users_organizations;

-- organizations.logo_object_key is the pre-files-table logo storage. Its data
-- was migrated into the files table (20251009T140000Z) and the application now
-- uses logo_file_id / horizontal_logo_file_id instead.
ALTER TABLE organizations
    DROP COLUMN logo_object_key;

-- The trust center NDA-acceptance flow now records acceptance through
-- electronic_signature_id together with the state column. The original boolean
-- flag, its metadata, the standalone NDA file reference, and the token expiry
-- column are all unused.
ALTER TABLE trust_center_accesses
    DROP COLUMN has_accepted_non_disclosure_agreement,
    DROP COLUMN has_accepted_non_disclosure_agreement_metadata,
    DROP COLUMN nda_file_id,
    DROP COLUMN last_token_expires_at;
