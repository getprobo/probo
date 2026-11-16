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

-- The hand-filled register keeps the table assets_registry and takes the
-- name RegistryAsset. Tables, columns, and constraints were renamed with
-- the register table. This renames the enum, the document link column,
-- and stored OAuth scopes.

ALTER TYPE asset_type RENAME TO registry_asset_type;

ALTER TABLE generated_documents
    RENAME COLUMN asset_list_document_id TO registry_asset_list_document_id;

UPDATE iam_oauth2_access_tokens
SET scopes = (
    SELECT array_agg(
        CASE scope
            WHEN 'v1:asset' THEN 'v1:registry-asset'
            WHEN 'v1:asset:read' THEN 'v1:registry-asset:read'
            ELSE scope
        END
        ORDER BY ordinality
    )
    FROM unnest(scopes) WITH ORDINALITY AS item(scope, ordinality)
)
WHERE scopes && ARRAY['v1:asset', 'v1:asset:read']::text[];

UPDATE iam_oauth2_refresh_tokens
SET scopes = (
    SELECT array_agg(
        CASE scope
            WHEN 'v1:asset' THEN 'v1:registry-asset'
            WHEN 'v1:asset:read' THEN 'v1:registry-asset:read'
            ELSE scope
        END
        ORDER BY ordinality
    )
    FROM unnest(scopes) WITH ORDINALITY AS item(scope, ordinality)
)
WHERE scopes && ARRAY['v1:asset', 'v1:asset:read']::text[];

UPDATE iam_oauth2_authorization_codes
SET scopes = (
    SELECT array_agg(
        CASE scope
            WHEN 'v1:asset' THEN 'v1:registry-asset'
            WHEN 'v1:asset:read' THEN 'v1:registry-asset:read'
            ELSE scope
        END
        ORDER BY ordinality
    )
    FROM unnest(scopes) WITH ORDINALITY AS item(scope, ordinality)
)
WHERE scopes && ARRAY['v1:asset', 'v1:asset:read']::text[];

UPDATE iam_oauth2_device_codes
SET scopes = (
    SELECT array_agg(
        CASE scope
            WHEN 'v1:asset' THEN 'v1:registry-asset'
            WHEN 'v1:asset:read' THEN 'v1:registry-asset:read'
            ELSE scope
        END
        ORDER BY ordinality
    )
    FROM unnest(scopes) WITH ORDINALITY AS item(scope, ordinality)
)
WHERE scopes && ARRAY['v1:asset', 'v1:asset:read']::text[];

UPDATE iam_oauth2_consents
SET scopes = (
    SELECT array_agg(
        CASE scope
            WHEN 'v1:asset' THEN 'v1:registry-asset'
            WHEN 'v1:asset:read' THEN 'v1:registry-asset:read'
            ELSE scope
        END
        ORDER BY ordinality
    )
    FROM unnest(scopes) WITH ORDINALITY AS item(scope, ordinality)
)
WHERE scopes && ARRAY['v1:asset', 'v1:asset:read']::text[];

UPDATE iam_oauth2_clients
SET scopes = (
    SELECT array_agg(
        CASE scope
            WHEN 'v1:asset' THEN 'v1:registry-asset'
            WHEN 'v1:asset:read' THEN 'v1:registry-asset:read'
            ELSE scope
        END
        ORDER BY ordinality
    )
    FROM unnest(scopes) WITH ORDINALITY AS item(scope, ordinality)
)
WHERE scopes && ARRAY['v1:asset', 'v1:asset:read']::text[];
