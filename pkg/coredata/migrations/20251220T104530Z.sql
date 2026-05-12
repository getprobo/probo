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
