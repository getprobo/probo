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

ALTER TABLE iam_membership_profiles
ADD COLUMN user_name TEXT,
ADD COLUMN external_id TEXT;

UPDATE iam_membership_profiles p
SET user_name = i.email_address
FROM identities i
WHERE i.id = p.identity_id AND p.source = 'SCIM';

CREATE UNIQUE INDEX idx_profiles_user_name_organization_id
ON iam_membership_profiles (user_name, organization_id)
WHERE user_name IS NOT NULL;

CREATE UNIQUE INDEX idx_profiles_external_id_organization_id
ON iam_membership_profiles (external_id, organization_id)
WHERE external_id IS NOT NULL;
