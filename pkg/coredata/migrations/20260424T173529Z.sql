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

CREATE TABLE agent_runs (
    id              TEXT NOT NULL PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    start_agent_name TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'PENDING',
    checkpoint      JSONB,
    input_messages  JSONB NOT NULL,
    result          JSONB,
    error_message   TEXT,
    started_at      TIMESTAMPTZ,
    lease_owner     TEXT,
    lease_expires_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agent_runs_running_lease ON agent_runs (lease_expires_at)
    WHERE status = 'RUNNING';
