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
