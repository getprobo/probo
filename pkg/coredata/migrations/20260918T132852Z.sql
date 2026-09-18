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

-- Give every connector-backed access review source an account to hang off,
-- so a source always names both the credential and the account it reviews.
--
-- The index swap and AccessReviewSource.Insert's conflict target ship in the
-- same release: Postgres infers a partial arbiter only when the conflict
-- target names exactly the index columns, so the old
-- ON CONFLICT (connector_id) against the new index raises 42P10 at plan
-- time for every source insert, CSV included.

ALTER TABLE access_review_sources
    ADD COLUMN connector_account_id TEXT REFERENCES connector_accounts(id)
        ON UPDATE CASCADE;

-- No foreign key, deliberately: the snapshot has to survive deletion of the
-- live account, exactly as its denormalized connector_id already does.
ALTER TABLE access_review_campaign_sources
    ADD COLUMN connector_account_id TEXT;

-- One account row per connector, carrying the settings-implied identifier
-- where the provider has one and NULL where the credential is the account.
-- The CASE mirrors Connector.ImpliedAccountID arm for arm.
--
-- Every connector gets a row, not only the sourced ones, so a backfilled
-- connector and a freshly created one are the same shape: otherwise adding
-- a source to a pre-existing SCIM-only connector would find no account to
-- attach to.
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
    generate_gid(decode_base64_unpadded(c.tenant_id), 134),
    c.tenant_id,
    c.organization_id,
    c.id,
    implied.external_account_id,
    COALESCE(implied.external_account_id, c.provider::text),
    NOW(),
    NOW()
FROM connectors c
CROSS JOIN LATERAL (
    SELECT NULLIF(
        CASE c.provider
            WHEN 'AWS'              THEN split_part(c.settings ->> 'role_arn', ':', 5)
            WHEN 'GCP'              THEN substring(c.settings ->> 'workload_identity_provider' FROM 'projects/([1-9][0-9]*)')
            WHEN 'AZURE'            THEN c.settings ->> 'subscription_id'
            WHEN 'GITHUB'           THEN c.settings ->> 'organization'
            WHEN 'SENTRY'           THEN c.settings ->> 'organization_slug'
            WHEN 'SUPABASE'         THEN c.settings ->> 'organization_slug'
            WHEN 'TALLY'            THEN c.settings ->> 'organization_id'
            WHEN 'QOVERY'           THEN c.settings ->> 'organization_id'
            WHEN 'NEON'             THEN c.settings ->> 'organization_id'
            WHEN 'SCALEWAY'         THEN c.settings ->> 'organization_id'
            WHEN 'GITLAB'           THEN c.settings ->> 'group_id'
            WHEN 'BITBUCKET'        THEN c.settings ->> 'workspace'
            WHEN 'ASANA'            THEN c.settings ->> 'workspace_gid'
            WHEN 'HEROKU'           THEN c.settings ->> 'team_id'
            WHEN 'CLICKUP'          THEN c.settings ->> 'team_id'
            WHEN 'VERCEL'           THEN c.settings ->> 'team_id'
            WHEN 'BETTER_STACK'     THEN c.settings ->> 'team_name'
            WHEN 'PAGERDUTY'        THEN c.settings ->> 'subdomain'
            WHEN 'ZENDESK'          THEN c.settings ->> 'subdomain'
            WHEN 'NETLIFY'          THEN c.settings ->> 'account_slug'
            WHEN 'DOCUSIGN'         THEN c.settings ->> 'account_id'
            WHEN 'CLOUDFLARE'       THEN c.settings ->> 'account_id'
            WHEN 'GOOGLE_ANALYTICS' THEN c.settings ->> 'account_id'
            WHEN 'ONE_PASSWORD'     THEN c.settings ->> 'account_id'
            WHEN 'OKTA'             THEN c.settings ->> 'domain'
            WHEN 'DATADOG'          THEN c.settings ->> 'domain'
            WHEN 'RENDER'           THEN c.settings ->> 'owner_id'
            WHEN 'CRISP'            THEN c.settings ->> 'website_id'
            WHEN 'TWINGATE'         THEN c.settings ->> 'network'
            ELSE NULL
        END,
        ''
    ) AS external_account_id
) implied
-- Defensive: in a fresh deployment the table was created empty one migration
-- ago and nothing can have written to it yet, so this is always true. It
-- keeps the backfill a no-op against a database where rows already exist,
-- which is what a shared development database looks like.
WHERE NOT EXISTS (
    SELECT 1 FROM connector_accounts a WHERE a.connector_id = c.id
);

UPDATE access_review_sources s
SET connector_account_id = a.id
FROM connector_accounts a
WHERE a.connector_id = s.connector_id
    AND s.connector_id IS NOT NULL;

UPDATE access_review_campaign_sources cs
SET connector_account_id = s.connector_account_id
FROM access_review_sources s
WHERE s.id = cs.access_review_source_id;

-- Two sources on one connector cannot both point at its single account row.
-- The unconditional idx_access_review_sources_connector_id already makes
-- that impossible, so reaching here is corruption: fail loudly rather than
-- picking a winner and silently reviewing one account twice.
DO $$
DECLARE
    orphan_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO orphan_count
    FROM access_review_sources
    WHERE connector_id IS NOT NULL AND connector_account_id IS NULL;

    IF orphan_count > 0 THEN
        RAISE EXCEPTION
            'cannot backfill access_review_sources.connector_account_id: % connector-backed source(s) got no account row; several sources share one connector',
            orphan_count;
    END IF;
END
$$;

-- The old index is unconditional on the connector and so would refuse a
-- second account source on one organization-wide credential.
DROP INDEX idx_access_review_sources_connector_id;

CREATE UNIQUE INDEX idx_access_review_sources_connector_account
    ON access_review_sources (connector_id, connector_account_id)
    NULLS NOT DISTINCT
    WHERE connector_id IS NOT NULL;

-- Validated, not NOT VALID: the backfill above is what makes it satisfiable,
-- and an unvalidated constraint is one nobody ever comes back to.
ALTER TABLE access_review_sources
    ADD CONSTRAINT access_review_sources_connector_account_pair
    CHECK ((connector_id IS NULL) = (connector_account_id IS NULL));
