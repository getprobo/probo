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

CREATE TYPE device_state AS ENUM (
    'PENDING',
    'ACTIVE',
    'REVOKED'
);

CREATE TABLE devices (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    state device_state NOT NULL DEFAULT 'PENDING',
    hardware_uuid TEXT,
    serial_number TEXT,
    hostname TEXT,
    platform device_platform,
    os_version TEXT,
    agent_version TEXT,
    api_key_hash BYTEA NOT NULL,
    owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    enrolled_at TIMESTAMP WITH TIME ZONE,
    last_seen_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT devices_active_fields_check CHECK (
        state != 'ACTIVE'
        OR (
            hardware_uuid IS NOT NULL
            AND hostname IS NOT NULL
            AND platform IS NOT NULL
            AND os_version IS NOT NULL
            AND agent_version IS NOT NULL
            AND enrolled_at IS NOT NULL
            AND last_seen_at IS NOT NULL
        )
    ),
    CONSTRAINT devices_revoked_at_check CHECK (
        (state = 'REVOKED' AND revoked_at IS NOT NULL)
        OR (state != 'REVOKED' AND revoked_at IS NULL)
    )
);

CREATE UNIQUE INDEX devices_org_hardware_uuid_idx
    ON devices (organization_id, hardware_uuid)
    WHERE hardware_uuid IS NOT NULL
      AND state != 'REVOKED';

CREATE UNIQUE INDEX devices_api_key_hash_idx
    ON devices (api_key_hash);

CREATE INDEX devices_owner_profile_idx
    ON devices (owner_profile_id)
    WHERE owner_profile_id IS NOT NULL;

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
