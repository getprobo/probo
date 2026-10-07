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

-- Platform accounts under a connector. A source that names a connector always
-- names one of these; CSV sources keep both sides null. Entity type 134 is
-- ConnectorAccountEntityType.

CREATE TABLE connector_accounts (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    connector_id TEXT NOT NULL REFERENCES connectors(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    external_account_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_connector_accounts_connector_external
    ON connector_accounts (connector_id, external_account_id);

ALTER TABLE access_review_sources
    ADD COLUMN connector_account_id TEXT
        REFERENCES connector_accounts(id)
        ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE access_review_campaign_sources
    ADD COLUMN connector_account_id TEXT;

-- Implied vendor account from connector settings. AWS and GCP parse the
-- identity out of a resource name; everyone else uses the settings key the
-- driver already treats as the tenant. One source per connector, so the new
-- account is linked by connector id.
WITH source_accounts AS (
    SELECT
        c.tenant_id,
        c.organization_id,
        c.id AS connector_id,
        COALESCE(
            CASE c.provider
                WHEN 'AWS' THEN substring(c.settings->>'role_arn' FROM 'iam::([0-9]+):')
                WHEN 'GCP' THEN substring(c.settings->>'workload_identity_provider' FROM 'projects/([1-9][0-9]*)/')
                WHEN 'AZURE' THEN c.settings->>'subscription_id'
                WHEN 'GITHUB' THEN c.settings->>'organization'
                WHEN 'SENTRY' THEN c.settings->>'organization_slug'
                WHEN 'GITLAB' THEN c.settings->>'group_id'
                WHEN 'BITBUCKET' THEN c.settings->>'workspace'
                WHEN 'HEROKU' THEN c.settings->>'team_id'
                WHEN 'ASANA' THEN c.settings->>'workspace_gid'
                WHEN 'NETLIFY' THEN c.settings->>'account_slug'
                WHEN 'CLICKUP' THEN c.settings->>'team_id'
                WHEN 'CLOUDFLARE' THEN c.settings->>'account_id'
                WHEN 'DOCUSIGN' THEN c.settings->>'account_id'
                WHEN 'PAGERDUTY' THEN c.settings->>'subdomain'
                WHEN 'VERCEL' THEN c.settings->>'team_id'
                WHEN 'DATADOG' THEN c.settings->>'domain'
                WHEN 'ZENDESK' THEN c.settings->>'subdomain'
                WHEN 'GOOGLE_ANALYTICS' THEN c.settings->>'account_id'
                WHEN 'OKTA' THEN c.settings->>'domain'
                WHEN 'SUPABASE' THEN c.settings->>'organization_slug'
                WHEN 'TALLY' THEN c.settings->>'organization_id'
                WHEN 'QOVERY' THEN c.settings->>'organization_id'
                WHEN 'NEON' THEN c.settings->>'organization_id'
                WHEN 'SCALEWAY' THEN c.settings->>'organization_id'
                WHEN 'CRISP' THEN c.settings->>'website_id'
                WHEN 'RENDER' THEN c.settings->>'owner_id'
                WHEN 'TWINGATE' THEN c.settings->>'network'
                WHEN 'BETTER_STACK' THEN c.settings->>'team_name'
                WHEN 'GRAFANA' THEN c.settings->>'base_url'
                WHEN 'SIGNOZ' THEN c.settings->>'base_url'
                WHEN 'METABASE' THEN c.settings->>'instance_url'
                WHEN 'POSTHOG' THEN c.settings->>'base_url'
                WHEN 'LANGFUSE' THEN c.settings->>'base_url'
                WHEN 'AUTHENTIK' THEN c.settings->>'base_url'
                WHEN 'SEGMENT' THEN c.settings->>'base_url'
                WHEN 'NEW_RELIC' THEN c.settings->>'region'
                WHEN 'RETOOL' THEN c.settings->>'base_url'
                WHEN 'ONE_PASSWORD' THEN COALESCE(c.settings->>'account_id', c.settings->>'scim_bridge_url')
                ELSE COALESCE(
                    c.settings->>'organization',
                    c.settings->>'organization_slug',
                    c.settings->>'organization_id',
                    c.settings->>'account_id',
                    c.settings->>'workspace',
                    c.settings->>'subdomain',
                    c.settings->>'domain',
                    c.settings->>'team_id',
                    c.settings->>'group_id'
                )
            END,
            ''
        ) AS external_account_id
    FROM access_review_sources s
    JOIN connectors c ON c.id = s.connector_id
    WHERE s.connector_id IS NOT NULL
),
inserted AS (
    INSERT INTO connector_accounts (
        id,
        tenant_id,
        organization_id,
        connector_id,
        external_account_id,
        name,
        created_at,
        updated_at
    )
    SELECT
        generate_gid(decode_base64_unpadded(tenant_id), 134),
        tenant_id,
        organization_id,
        connector_id,
        external_account_id,
        external_account_id,
        NOW(),
        NOW()
    FROM source_accounts
    RETURNING id, connector_id
)
UPDATE access_review_sources s
SET connector_account_id = inserted.id
FROM inserted
WHERE s.connector_id = inserted.connector_id;

UPDATE access_review_campaign_sources cs
SET connector_account_id = s.connector_account_id
FROM access_review_sources s
WHERE cs.access_review_source_id = s.id
    AND s.connector_account_id IS NOT NULL;

DROP INDEX idx_access_review_sources_connector_id;

CREATE UNIQUE INDEX idx_access_review_sources_connector_account
    ON access_review_sources (connector_id, connector_account_id)
    NULLS NOT DISTINCT
    WHERE connector_id IS NOT NULL;

ALTER TABLE access_review_sources
    ADD CONSTRAINT access_review_sources_connector_account_pair
    CHECK ((connector_id IS NULL) = (connector_account_id IS NULL));

ALTER TABLE access_review_sources
    VALIDATE CONSTRAINT access_review_sources_connector_account_pair;
