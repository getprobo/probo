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
