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
