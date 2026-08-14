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

ALTER TABLE probot_identity_bindings
    ADD COLUMN external_tenant_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN external_user_name TEXT NOT NULL DEFAULT '';

UPDATE probot_identity_bindings AS binding
SET
    external_tenant_name = challenge.external_tenant_name,
    external_user_name = challenge.external_user_name
FROM (
    SELECT DISTINCT ON (provider, external_tenant_id, external_user_id)
        provider,
        external_tenant_id,
        external_user_id,
        external_tenant_name,
        external_user_name
    FROM probot_identity_binding_challenges
    ORDER BY
        provider,
        external_tenant_id,
        external_user_id,
        created_at DESC
) AS challenge
WHERE
    binding.provider = challenge.provider
    AND binding.external_tenant_id = challenge.external_tenant_id
    AND binding.external_user_id = challenge.external_user_id;

ALTER TABLE probot_identity_bindings
    ALTER COLUMN external_tenant_name DROP DEFAULT,
    ALTER COLUMN external_user_name DROP DEFAULT;
