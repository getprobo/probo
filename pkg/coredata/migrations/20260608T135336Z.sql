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

-- Backfill organization_id on logo files that were inserted with the nil GID.
-- Before the fix in CreateOrganization/UpdateOrganization, the OrganizationID
-- field was omitted from the File struct, causing the nil GID
-- (AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA) to be stored instead of the actual org ID.

UPDATE files
SET organization_id = o.id
FROM organizations o
WHERE files.organization_id = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
  AND (
    files.id = o.logo_file_id
    OR files.id = o.horizontal_logo_file_id
  );

UPDATE files
SET organization_id = tc.organization_id
FROM trust_centers tc
WHERE files.organization_id = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
  AND (
    files.id = tc.logo_file_id
    OR files.id = tc.dark_logo_file_id
  );

UPDATE files
SET organization_id = tcr.organization_id
FROM trust_center_references tcr
WHERE files.organization_id = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
  AND files.id = tcr.logo_file_id;

UPDATE files
SET organization_id = f.organization_id
FROM frameworks f
WHERE files.organization_id = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
  AND (
    files.id = f.light_logo_file_id
    OR files.id = f.dark_logo_file_id
  );
