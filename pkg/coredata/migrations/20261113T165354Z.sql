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

-- A device row is created either by the Probo agent enrolling (AGENT) or by
-- a device sync from an MDM connector (MDM). The source decides which
-- lifecycle governs the row: the agent's enrollment and heartbeats, or the
-- sync, which activates the row while the MDM lists the device and revokes
-- it once the MDM stops listing it.
--
-- An agent device whose serial number an MDM also lists keeps source AGENT
-- and gains the mdm_* columns, so one machine is one row. mdm_provider is
-- copied from the connector so that a row still names its MDM once the
-- connector is deleted and mdm_connector_id is cleared.
CREATE TYPE device_source AS ENUM (
    'AGENT',
    'MDM'
);

ALTER TABLE devices
    ADD COLUMN source device_source NOT NULL DEFAULT 'AGENT',
    ADD COLUMN model TEXT,
    ADD COLUMN mdm_provider connector_provider,
    ADD COLUMN mdm_connector_id TEXT REFERENCES connectors(id) ON UPDATE CASCADE ON DELETE SET NULL,
    ADD COLUMN mdm_device_id TEXT,
    ADD COLUMN mdm_last_check_in_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE devices
    ALTER COLUMN source DROP DEFAULT;

-- The agent-specific invariants now hold for agent rows only. A synced row
-- is ACTIVE without an agent version, enrollment or API key.
ALTER TABLE devices
    DROP CONSTRAINT devices_active_fields_check;

ALTER TABLE devices
    ADD CONSTRAINT devices_active_fields_check CHECK (
        state != 'ACTIVE'
        OR source != 'AGENT'
        OR (
            hardware_uuid IS NOT NULL
            AND hostname IS NOT NULL
            AND platform IS NOT NULL
            AND os_version IS NOT NULL
            AND agent_version IS NOT NULL
            AND enrolled_at IS NOT NULL
            AND last_seen_at IS NOT NULL
        )
    );

ALTER TABLE devices
    DROP CONSTRAINT devices_active_api_key_hash_check;

ALTER TABLE devices
    ADD CONSTRAINT devices_active_api_key_hash_check CHECK (
        state != 'ACTIVE'
        OR source != 'AGENT'
        OR api_key_hash IS NOT NULL
    );

-- A synced row never holds an agent credential, so it can never
-- authenticate against the agent API.
ALTER TABLE devices
    ADD CONSTRAINT devices_mdm_no_api_key_check CHECK (
        source != 'MDM'
        OR api_key_hash IS NULL
    );

ALTER TABLE devices
    ADD CONSTRAINT devices_mdm_device_id_check CHECK (
        source != 'MDM'
        OR (mdm_device_id IS NOT NULL AND mdm_provider IS NOT NULL)
    );

CREATE UNIQUE INDEX devices_mdm_device_idx
    ON devices (mdm_connector_id, mdm_device_id)
    WHERE mdm_connector_id IS NOT NULL
      AND deleted_at IS NULL;
