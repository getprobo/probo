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

CREATE TYPE connector_protocol AS ENUM ('OAUTH2');
CREATE TYPE connector_provider AS ENUM ('SLACK');

ALTER TABLE connectors DROP COLUMN type;
ALTER TABLE connectors DROP COLUMN name;

ALTER TABLE connectors ADD COLUMN protocol connector_protocol NOT NULL;
ALTER TABLE connectors ADD COLUMN provider connector_provider NOT NULL;
ALTER TABLE connectors ADD COLUMN settings JSONB;

DROP INDEX IF EXISTS idx_connectors_organization_id_name;

CREATE TABLE slack_messages (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	organization_id TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	sent_at TIMESTAMP WITH TIME ZONE,
	error TEXT,
	CONSTRAINT fk_slack_messages_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX ON slack_messages (sent_at) WHERE sent_at IS NULL AND error IS NULL;
