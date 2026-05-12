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
