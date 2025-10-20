CREATE TYPE slack_message_type AS ENUM ('TRUST_CENTER_ACCESS_REQUEST', 'WELCOME');

ALTER TABLE slack_messages ALTER COLUMN body TYPE JSONB USING body::jsonb;
ALTER TABLE slack_messages ADD COLUMN message_ts TEXT;
ALTER TABLE slack_messages ADD COLUMN channel_id TEXT;
ALTER TABLE slack_messages ADD COLUMN requester_email TEXT;
ALTER TABLE slack_messages ADD COLUMN type slack_message_type NOT NULL;

ALTER TABLE slack_messages ADD CONSTRAINT unique_slack_message_org_ts_channel UNIQUE (organization_id, message_ts, channel_id);

CREATE TABLE slack_message_updates (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	slack_message_id TEXT NOT NULL,
	body JSONB NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	sent_at TIMESTAMP WITH TIME ZONE,
	error TEXT,
	CONSTRAINT fk_slack_message_updates_slack_message_id FOREIGN KEY (slack_message_id) REFERENCES slack_messages(id) ON UPDATE CASCADE ON DELETE CASCADE
);
