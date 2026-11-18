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

CREATE TYPE agent_execution_kind AS ENUM (
    'ONE_SHOT',
    'CONVERSATIONAL'
);

ALTER TABLE agent_runs
    ADD COLUMN execution_kind agent_execution_kind NOT NULL DEFAULT 'ONE_SHOT',
    ADD COLUMN source TEXT,
    ADD COLUMN session_key TEXT,
    ADD COLUMN source_coordinates JSONB,
    ADD COLUMN trusted_context JSONB,
    ADD COLUMN session_messages JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN processing_owner_token TEXT,
    ADD COLUMN processing_heartbeat_at TIMESTAMPTZ,
    ADD COLUMN processing_input_ids TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN attempt_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN next_attempt_at TIMESTAMPTZ,
    ADD COLUMN last_error TEXT,
    ADD COLUMN dead_lettered_at TIMESTAMPTZ;

UPDATE agent_runs
SET session_messages = input_messages;

ALTER TABLE agent_runs
    ALTER COLUMN execution_kind DROP DEFAULT,
    ALTER COLUMN session_messages DROP DEFAULT,
    ALTER COLUMN processing_input_ids DROP DEFAULT,
    ALTER COLUMN attempt_count DROP DEFAULT,
    ALTER COLUMN max_attempts DROP DEFAULT;

CREATE UNIQUE INDEX agent_runs_source_session_key
    ON agent_runs (tenant_id, organization_id, source, session_key)
    WHERE source IS NOT NULL AND session_key IS NOT NULL;

CREATE TABLE agent_inputs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    agent_run_id TEXT NOT NULL,
    source TEXT NOT NULL,
    source_event_id TEXT,
    message JSONB NOT NULL,
    processed_at TIMESTAMPTZ,
    attempt_count INTEGER NOT NULL,
    max_attempts INTEGER NOT NULL,
    next_attempt_at TIMESTAMPTZ,
    last_error TEXT,
    dead_lettered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT agent_inputs_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT agent_inputs_agent_run_id_fkey
        FOREIGN KEY (agent_run_id) REFERENCES agent_runs(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX agent_inputs_source_event_key
    ON agent_inputs (tenant_id, organization_id, source, source_event_id)
    WHERE source_event_id IS NOT NULL;
