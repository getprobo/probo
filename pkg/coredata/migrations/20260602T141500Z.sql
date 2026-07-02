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
