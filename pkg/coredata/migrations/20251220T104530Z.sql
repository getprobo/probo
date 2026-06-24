-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

ALTER TABLE users RENAME TO identities;
ALTER TABLE sessions RENAME COLUMN user_id TO identity_id;
ALTER TABLE authz_memberships RENAME COLUMN user_id TO identity_id;
ALTER TABLE peoples RENAME COLUMN user_id TO identity_id;
ALTER TABLE auth_user_api_keys RENAME COLUMN user_id TO identity_id;
ALTER TABLE authz_api_keys_memberships RENAME COLUMN auth_user_api_key_id TO personal_api_key_id;
ALTER TABLE sessions RENAME TO iam_sessions;
ALTER TABLE authz_memberships RENAME TO iam_memberships;
ALTER TABLE authz_invitations RENAME TO iam_invitations;
ALTER TABLE auth_user_api_keys RENAME TO iam_personal_api_keys;
ALTER TABLE authz_api_keys_memberships RENAME TO iam_personal_api_key_memberships;
ALTER TABLE auth_saml_configurations RENAME TO iam_saml_configurations;
ALTER TABLE auth_saml_assertions RENAME TO iam_saml_assertions;
ALTER TABLE auth_saml_requests RENAME TO iam_saml_requests;
