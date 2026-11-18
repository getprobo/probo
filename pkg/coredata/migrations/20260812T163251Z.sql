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

CREATE TYPE slackbot_installation_status AS ENUM (
    'ACTIVE',
    'DISABLED'
);

CREATE TABLE slackbot_installations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL UNIQUE REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    team_id TEXT NOT NULL UNIQUE,
    bot_user_id TEXT NOT NULL,
    encrypted_access_token BYTEA NOT NULL,
    encrypted_refresh_token BYTEA,
    access_token_expires_at TIMESTAMP WITH TIME ZONE,
    scopes TEXT[] NOT NULL,
    status slackbot_installation_status NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE slackbot_messages (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    message_type TEXT NOT NULL,
    body JSONB NOT NULL,
    metadata JSONB NOT NULL,
    channel_id TEXT,
    message_ts TEXT,
    initial_slackbot_message_id TEXT NOT NULL REFERENCES slackbot_messages(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    processing_started_at TIMESTAMP WITH TIME ZONE,
    attempt_count INTEGER NOT NULL,
    max_attempts INTEGER NOT NULL,
    last_attempted_at TIMESTAMP WITH TIME ZONE,
    next_attempt_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    client_msg_id UUID NOT NULL,
    error TEXT
);

CREATE TABLE bot_delivery_destinations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    provider TEXT NOT NULL,
    target_namespace TEXT NOT NULL,
    target_key TEXT NOT NULL,
    external_destination_id TEXT NOT NULL,
    external_name TEXT NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (provider, target_namespace, target_key)
);

CREATE TABLE slackbot_interactive_claims (
    interaction_digest BYTEA PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    processing_token UUID,
    processing_started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE slackbot_install_state_claims (
    state_digest BYTEA PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    processing_token UUID,
    processing_started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE slackbot_events (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL UNIQUE,
    envelope JSONB NOT NULL,
    processing_started_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    attempt_count INTEGER NOT NULL,
    max_attempts INTEGER NOT NULL,
    next_attempt_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    dead_lettered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
