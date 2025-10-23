CREATE TYPE slack_message_type AS ENUM ('TRUST_CENTER_ACCESS_REQUEST', 'WELCOME');

ALTER TABLE slack_messages ALTER COLUMN body TYPE JSONB USING body::jsonb;
ALTER TABLE slack_messages ADD COLUMN message_ts TEXT;
ALTER TABLE slack_messages ADD COLUMN channel_id TEXT;
ALTER TABLE slack_messages ADD COLUMN requester_email TEXT;
ALTER TABLE slack_messages ADD COLUMN type slack_message_type NOT NULL;
ALTER TABLE slack_messages ADD COLUMN metadata JSONB;
ALTER TABLE slack_messages ADD COLUMN initial_slack_message_id TEXT NOT NULL;
ALTER TABLE slack_messages ADD CONSTRAINT fk_slack_messages_initial_slack_message_id
    FOREIGN KEY (initial_slack_message_id) REFERENCES slack_messages(id) ON DELETE CASCADE;
