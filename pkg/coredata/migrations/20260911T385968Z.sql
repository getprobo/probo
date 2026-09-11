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

-- A vendor app-install callback is a public GET whose proof is long-lived
-- (unlike an OAuth authorization code, which the vendor itself burns). Nothing
-- else stops a browser refresh, a back-button, or a link preview prefetch from
-- executing it twice: the (organization_id, provider, protocol) unique index on
-- connectors was deliberately dropped in 20260819T142937Z, and there is no HTTP
-- rate limiting anywhere in this repository. This table is the single-use ledger
-- that makes one signed state usable exactly once.
--
-- It dedupes a STATE, not a website: the primary key is the state digest and
-- every mint carries a fresh nonce, so two concurrent tabs are two rows and both
-- claims succeed. Serializing two first-installs of the same website is a
-- separate mechanism -- an advisory lock inside CompleteInstall's transaction --
-- because it has to hold over a row that does not exist yet.
--
-- THIS IS THE LEDGER FOR EVERY CONNECTOR APP-INSTALL CEREMONY. A provider
-- gaining an Install block gets single-use semantics here, for free and without
-- a migration: the service layer claims and burns through
-- coredata.NewConnectorInstallStateClaim on every install provider, so there is
-- no seam at which a new one could end up with a ledger of its own.
--
-- slackbot_install_state_claims (20260812T163251Z) is structurally identical and
-- stays separate on purpose. Slack's install is its own feature with its own
-- lifecycle and its own retention; the two share the coredata.InstallStateClaim
-- mechanism but never a table, and merging them is explicitly not a goal.
--
-- state_digest is a SHA-256 of the whole signed state token, so the raw state --
-- which carries the organization and identity GIDs -- is never at rest here.
CREATE TABLE connector_install_state_claims (
    state_digest BYTEA PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    processing_token UUID,
    processing_started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- No index on created_at: nothing sweeps this table yet, and slackbot's twin
-- ships without one. It belongs in the migration that adds the sweep, shaped to
-- that sweep's predicate.
