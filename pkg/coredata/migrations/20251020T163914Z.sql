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
