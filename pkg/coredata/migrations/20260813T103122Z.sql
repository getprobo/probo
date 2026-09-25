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

CREATE TYPE bot_message_purpose AS ENUM (
    'POST',
    'UPDATE'
);

CREATE TABLE bot_messages (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    capability TEXT NOT NULL,
    message_type TEXT NOT NULL,
    attributes JSONB NOT NULL,
    subject_namespace TEXT NOT NULL,
    subject_key TEXT NOT NULL,
    event_key TEXT NOT NULL,
    purpose bot_message_purpose NOT NULL,
    processing_started_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ,
    attempt_count INTEGER NOT NULL,
    max_attempts INTEGER NOT NULL,
    next_attempt_at TIMESTAMPTZ,
    last_error TEXT,
    dead_lettered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT bot_messages_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX bot_messages_event_key
    ON bot_messages (
        tenant_id,
        organization_id,
        subject_namespace,
        subject_key,
        event_key
    );

CREATE TABLE bot_thread_subjects (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    external_conversation_id TEXT NOT NULL,
    external_message_id TEXT NOT NULL,
    capability TEXT NOT NULL,
    message_type TEXT NOT NULL,
    attributes JSONB NOT NULL,
    subject_namespace TEXT NOT NULL,
    subject_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT bot_thread_subjects_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX bot_thread_subjects_provider_coordinates
    ON bot_thread_subjects (
        tenant_id,
        organization_id,
        provider,
        external_conversation_id,
        external_message_id
    );

CREATE UNIQUE INDEX bot_thread_subjects_subject
    ON bot_thread_subjects (
        tenant_id,
        organization_id,
        subject_namespace,
        subject_key
    );
