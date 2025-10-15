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
