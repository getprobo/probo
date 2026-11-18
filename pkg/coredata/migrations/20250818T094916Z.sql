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

CREATE TYPE compliance_registries_status AS ENUM (
    'OPEN',
    'IN_PROGRESS',
    'CLOSED'
);

CREATE TABLE compliance_registries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    area TEXT,
    source TEXT,
    audit_id TEXT NOT NULL,
    requirement TEXT,
    actions_to_be_implemented TEXT,
    regulator TEXT,
    owner_id TEXT NOT NULL,
    last_review_date DATE,
    due_date DATE,
    status compliance_registries_status NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT compliance_registries_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT compliance_registries_owner_id_fkey
        FOREIGN KEY (owner_id)
        REFERENCES peoples(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    CONSTRAINT compliance_registries_audit_id_fkey
        FOREIGN KEY (audit_id)
        REFERENCES audits(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);
