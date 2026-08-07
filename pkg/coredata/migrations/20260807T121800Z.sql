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

CREATE TYPE ai_system_statuses AS ENUM (
    'ACTIVE',
    'IN_DEVELOPMENT',
    'DECOMMISSIONED'
);

CREATE TYPE ai_system_risk_classifications AS ENUM (
    'HIGH_RISK',
    'LIMITED',
    'MINIMAL',
    'GPAI'
);

CREATE TABLE ai_systems (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL,
    version TEXT,
    company_roles TEXT[] NOT NULL,
    status ai_system_statuses NOT NULL,
    owner_id TEXT REFERENCES iam_membership_profiles(id),
    source TEXT,
    purpose TEXT,
    intended_use_cases TEXT,
    autonomy_level TEXT,
    human_oversight_mechanism TEXT,
    risk_classification ai_system_risk_classifications,
    key_stakeholders TEXT,
    data_sources_and_type TEXT,
    deployment_date DATE,
    last_review_date DATE,
    next_review_date DATE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

ALTER TABLE generated_documents
    ADD COLUMN ai_systems_document_id TEXT REFERENCES documents(id) ON DELETE SET NULL;

-- Allow the prb CLI OAuth2 client to request the ai-system scope.
-- Keep in sync with pkg/cli/config.CLIClientScopes.
UPDATE iam_oauth2_clients
SET scopes = '{
    openid,
    profile,
    email,
    offline_access,
    v1:access-review,
    v1:agent,
    v1:ai-system,
    v1:asset,
    v1:audit,
    v1:business-function,
    v1:common-third-party,
    v1:compliance-page,
    v1:connector,
    v1:control,
    v1:datum,
    v1:document,
    v1:iam,
    v1:itam,
    v1:org,
    v1:privacy,
    v1:risk,
    v1:slack-connection,
    v1:task,
    v1:third-party,
    v1:webhook
}'::TEXT[],
    updated_at = NOW()
WHERE id = 'AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp';
