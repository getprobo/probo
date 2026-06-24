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

CREATE TABLE audit_log_entries (
    id              TEXT        PRIMARY KEY,
    tenant_id       TEXT        NOT NULL,
    organization_id TEXT        NOT NULL REFERENCES organizations(id),
    actor_id        TEXT        NOT NULL,
    actor_type      TEXT        NOT NULL,
    action          TEXT        NOT NULL,
    resource_type   TEXT        NOT NULL,
    resource_id     TEXT        NOT NULL,
    metadata        JSONB       NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_audit_log_entries_organization_id ON audit_log_entries (organization_id);
CREATE INDEX idx_audit_log_entries_actor_id ON audit_log_entries (actor_id);
CREATE INDEX idx_audit_log_entries_action ON audit_log_entries (action);
CREATE INDEX idx_audit_log_entries_created_at ON audit_log_entries (created_at);
