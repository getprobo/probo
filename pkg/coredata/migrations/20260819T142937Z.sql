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

-- Allow multiple connectors per (organization, provider) so an organization
-- can review access across several accounts of one provider (e.g. two GitHub
-- organizations, two Slack workspaces). The Slack messaging worker used to
-- rely on this uniqueness to resolve the organization's credential by
-- provider alone; it now picks deterministically (channel-configured first,
-- then oldest), so no provider keeps the cap.

DROP INDEX idx_connectors_organization_id_provider;

-- The dropped unique index was also the table's only organization-leading
-- index; keep a plain one for the per-organization lookups.
CREATE INDEX idx_connectors_organization_id_provider_lookup
    ON connectors (organization_id, provider);

-- Schema-forward-only: once intentional duplicates exist, the unique index
-- cannot be recreated without a product-level merge policy. Application
-- rollback keeps working (old binaries simply reconnect via their initiate
-- fallback); schema rollback does not.
