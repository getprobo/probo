-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

CREATE INDEX idx_agent_runs_status ON agent_runs (status)
    WHERE status IN ('PENDING', 'RUNNING', 'SUSPENDED');
CREATE INDEX idx_agent_runs_organization_status ON agent_runs (organization_id, status, created_at);
CREATE INDEX idx_agent_runs_running_lease ON agent_runs (lease_expires_at)
    WHERE status = 'RUNNING';
