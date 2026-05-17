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

CREATE TYPE device_platform AS ENUM (
    'DARWIN',
    'LINUX',
    'FREEBSD',
    'WINDOWS'
);

CREATE TYPE device_posture_status AS ENUM (
    'PASS',
    'FAIL',
    'UNKNOWN',
    'NOT_APPLICABLE'
);

CREATE TABLE device_enrollment_tokens (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL,
    token_hash BYTEA NOT NULL,
    created_by_identity_id TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    max_uses INTEGER,
    used_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX device_enrollment_tokens_org_idx
    ON device_enrollment_tokens (organization_id);

CREATE UNIQUE INDEX device_enrollment_tokens_token_hash_idx
    ON device_enrollment_tokens (token_hash);

CREATE TABLE devices (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    hardware_uuid TEXT NOT NULL,
    serial_number TEXT,
    hostname TEXT NOT NULL,
    platform device_platform NOT NULL,
    os_version TEXT NOT NULL,
    agent_version TEXT NOT NULL,
    api_key_hash BYTEA NOT NULL,
    assigned_user_identity_id TEXT,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    enrolled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX devices_org_hardware_uuid_idx
    ON devices (organization_id, hardware_uuid);

CREATE UNIQUE INDEX devices_api_key_hash_idx
    ON devices (api_key_hash);

CREATE INDEX devices_assigned_user_idx
    ON devices (assigned_user_identity_id)
    WHERE assigned_user_identity_id IS NOT NULL;

CREATE TABLE device_postures (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    device_id TEXT NOT NULL REFERENCES devices(id) ON UPDATE CASCADE ON DELETE CASCADE,
    check_key TEXT NOT NULL,
    status device_posture_status NOT NULL,
    evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    observed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX device_postures_device_id_check_key_observed_at_idx
    ON device_postures (device_id, check_key, observed_at DESC);

CREATE INDEX device_postures_organization_id_idx
    ON device_postures (organization_id);
