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

CREATE TYPE continual_improvement_registries_status AS ENUM (
    'OPEN',
    'IN_PROGRESS',
    'CLOSED'
);

CREATE TYPE continual_improvement_registries_priority AS ENUM (
    'LOW',
    'MEDIUM',
    'HIGH'
);

CREATE TABLE continual_improvement_registries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    description TEXT,
    audit_id TEXT NOT NULL,
    source TEXT,
    owner_id TEXT NOT NULL,
    target_date DATE,
    status continual_improvement_registries_status NOT NULL,
    priority continual_improvement_registries_priority NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT continual_improvement_registries_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT continual_improvement_registries_owner_id_fkey
        FOREIGN KEY (owner_id)
        REFERENCES peoples(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    CONSTRAINT continual_improvement_registries_audit_id_fkey
        FOREIGN KEY (audit_id)
        REFERENCES audits(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);
