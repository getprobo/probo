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

CREATE TABLE cloud_accounts (
    id                         TEXT NOT NULL PRIMARY KEY,
    tenant_id                  TEXT NOT NULL,
    organization_id            TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    label                      TEXT NOT NULL,
    provider                   TEXT NOT NULL,                   -- AWS | GCP | AZURE
    credential_kind            TEXT NOT NULL,                   -- AWS_ASSUME_ROLE | GCP_SERVICE_ACCOUNT_KEY | AZURE_CLIENT_SECRET | ...
    scope_kind                 TEXT NOT NULL,                   -- AWS_ACCOUNT | GCP_PROJECT | GCP_ORGANIZATION | AZURE_SUBSCRIPTION | AZURE_MANAGEMENT_GROUP | AZURE_TENANT | ...
    scope_identifier           TEXT NOT NULL,                   -- account id / project id / org id / mg id / subscription id / tenant id
    enabled_audit_modules      TEXT[] NOT NULL DEFAULT '{}',    -- ACCESS_REVIEW, CSPM_S3_PUBLIC, CSPM_IAM_AUDIT (v2 placeholders)
    status                     TEXT NOT NULL,                   -- PENDING_VERIFICATION | VERIFIED | ERRORED | DISCONNECTED
    consecutive_probe_failures INT NOT NULL DEFAULT 0,
    first_probe_failure_at     TIMESTAMP WITH TIME ZONE NULL,
    encrypted_credentials      BYTEA NOT NULL,                  -- prefix byte 0x01 = key version v1; future rotation can write 0x02 etc.
    external_id                TEXT NULL,                       -- AWS only; non-null for AWS_ASSUME_ROLE; 64 hex chars from crypto/rand
    template_sha256            TEXT NULL,                       -- AWS only; SHA-256 of the CFN template the customer's stack is pinned to (drift detection at rotate time)
    last_probe_at              TIMESTAMP WITH TIME ZONE NULL,
    last_probe_error           TEXT NULL,
    last_verified_at           TIMESTAMP WITH TIME ZONE NULL,
    created_at                 TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                 TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT cloud_accounts_org_label_unique UNIQUE (organization_id, label)
);
