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

-- Record which vendor accounts a connector credential covers. For a
-- standalone or SaaS connector that is exactly one row; for an org-wide
-- cloud credential it is one row per enabled member account.
--
-- external_account_id is NULL when the credential *is* the account: most
-- SaaS providers store no tenant identifier at all, and a NULL resolves to
-- the settings-implied account when a session is opened. It is never the
-- empty string, which would read as a slug we failed to capture.
--
-- ON DELETE CASCADE because accounts are owned by the credential and have no
-- life without it. Access review, SCIM and Settings all delete connectors;
-- none of them should have to sweep accounts first.
CREATE TABLE connector_accounts (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    connector_id TEXT NOT NULL REFERENCES connectors(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    external_account_id TEXT,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT connector_accounts_external_id_not_blank
        CHECK (external_account_id IS NULL OR external_account_id <> '')
);

-- One row per real vendor account ...
CREATE UNIQUE INDEX idx_connector_accounts_connector_external_id
    ON connector_accounts (connector_id, external_account_id);

-- ... and at most one "the credential is the account" row per connector.
-- Written as a second partial index rather than NULLS NOT DISTINCT so this
-- table's uniqueness does not depend on the behaviour gate 10.5 probes.
CREATE UNIQUE INDEX idx_connector_accounts_connector_implicit
    ON connector_accounts (connector_id)
    WHERE external_account_id IS NULL;
