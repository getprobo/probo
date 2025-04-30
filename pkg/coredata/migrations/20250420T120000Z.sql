CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- PostgreSQL implementation of the GID generation system

-- 1. First, create functions for TenantID generation
CREATE OR REPLACE FUNCTION generate_machine_id()
RETURNS bytea AS $func$
DECLARE
    machine_id bytea;
BEGIN
    -- Generate 3 random bytes for machine ID
    machine_id := decode(encode(gen_random_bytes(3), 'hex'), 'hex');
    RETURN machine_id;
END;
$func$ LANGUAGE plpgsql STABLE;

-- Store machine ID as a database-wide setting (run once)
DO $block$
BEGIN
    -- Check if the setting exists
    IF NOT EXISTS (SELECT 1 FROM pg_settings WHERE name = 'app.machine_id') THEN
        -- Create custom parameter in postgresql.conf or via ALTER SYSTEM
        PERFORM set_config('app.machine_id', encode(generate_machine_id(), 'hex'), false);
    END IF;
END $block$;

-- Counter for tenant ID generation (used atomically)
CREATE SEQUENCE IF NOT EXISTS tenant_id_counter_seq;

-- Function to generate a TenantID
CREATE OR REPLACE FUNCTION generate_tenant_id()
RETURNS bytea AS $func$
DECLARE
    id bytea;
    machine_id bytea;
    timestamp_bytes bytea;
    counter_bytes bytea;
    counter_val int;
BEGIN
    -- 1. Get machine ID (3 bytes)
    machine_id := decode(current_setting('app.machine_id'), 'hex');
    
    -- 2. Get timestamp bytes (3 bytes - Unix time in seconds)
    timestamp_bytes := substring(int8send(extract(epoch from now())::bigint) from 6 for 3);
    
    -- 3. Get counter (2 bytes)
    counter_val := nextval('tenant_id_counter_seq') % 65536; -- 2^16
    counter_bytes := substring(int4send(counter_val) from 3 for 2);
    
    -- 4. Combine all parts
    id := machine_id || timestamp_bytes || counter_bytes;
    
    RETURN id;
END;
$func$ LANGUAGE plpgsql VOLATILE;

-- Function to convert TenantID to string
CREATE OR REPLACE FUNCTION tenant_id_to_string(tenant_id bytea)
RETURNS text AS $func$
BEGIN
    -- Make sure to handle padding properly - remove trailing '=' characters
    RETURN rtrim(translate(encode(tenant_id, 'base64'), '+/', '-_'), '=');
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

-- Function to parse tenant ID from string
CREATE OR REPLACE FUNCTION parse_tenant_id(encoded text)
RETURNS bytea AS $func$
DECLARE
    decoded bytea;
    padded_input text;
    padding_needed int;
BEGIN
    -- Add proper padding for base64 decoding
    padding_needed := (4 - (length(encoded) % 4)) % 4;
    padded_input := encoded || repeat('=', padding_needed);
    
    -- Replace URL-safe chars and decode
    BEGIN
        decoded := decode(translate(padded_input, '-_', '+/'), 'base64');
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'Invalid base64 encoding in tenant ID';
    END;
    
    -- Validate length
    IF octet_length(decoded) != 8 THEN
        RAISE EXCEPTION 'Invalid tenant ID length: got %, want 8', octet_length(decoded);
    END IF;
    
    RETURN decoded;
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

-- 2. Now create functions for full GID generation

-- Function to generate a GID
CREATE OR REPLACE FUNCTION generate_gid(tenant_id bytea, entity_type int)
RETURNS text AS $func$
DECLARE
    id bytea;
    timestamp_ms_bytes bytea;
    entity_type_bytes bytea;
    random_bytes bytea;
BEGIN
    -- Validate tenant_id
    IF tenant_id IS NULL OR octet_length(tenant_id) != 8 THEN
        RAISE EXCEPTION 'Invalid tenant ID: must be 8 bytes';
    END IF;

    -- 1. Start with tenant ID (8 bytes)
    id := tenant_id;
    
    -- 2. Add entity type (2 bytes)
    entity_type_bytes := substring(int4send(entity_type) from 3 for 2);
    id := id || entity_type_bytes;
    
    -- 3. Add timestamp in milliseconds (8 bytes)
    -- Extract milliseconds since epoch
    timestamp_ms_bytes := int8send(
        (extract(epoch from now()) * 1000)::bigint
    );
    id := id || timestamp_ms_bytes;
    
    -- 4. Add random bytes for uniqueness (6 bytes)
    random_bytes := gen_random_bytes(6);
    id := id || random_bytes;
    
    -- Return as URL-safe base64
    RETURN rtrim(translate(encode(id, 'base64'), '+/', '-_'), '=');
END;
$func$ LANGUAGE plpgsql VOLATILE;

-- Note: removed the single-parameter overload - tenant_id must be explicitly provided

-- Function to convert GID to string
CREATE OR REPLACE FUNCTION gid_to_string(gid bytea)
RETURNS text AS $func$
BEGIN
    -- Make sure to handle padding properly - remove trailing '=' characters
    RETURN rtrim(translate(encode(gid, 'base64'), '+/', '-_'), '=');
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

-- Function to parse GID from string
CREATE OR REPLACE FUNCTION parse_gid(encoded text)
RETURNS bytea AS $func$
DECLARE
    decoded bytea;
    padded_input text;
    padding_needed int;
BEGIN
    -- Add proper padding for base64 decoding
    padding_needed := (4 - (length(encoded) % 4)) % 4;
    padded_input := encoded || repeat('=', padding_needed);
    
    -- Replace URL-safe chars and decode
    BEGIN
        decoded := decode(translate(padded_input, '-_', '+/'), 'base64');
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'Invalid base64 encoding in GID';
    END;
    
    -- Validate length
    IF octet_length(decoded) != 24 THEN
        RAISE EXCEPTION 'Invalid GID length: got %, want 24', octet_length(decoded);
    END IF;
    
    RETURN decoded;
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

-- Extract tenant ID from GID
CREATE OR REPLACE FUNCTION extract_tenant_id(gid bytea)
RETURNS bytea AS $func$
BEGIN
    RETURN substring(gid from 1 for 8);
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

-- Extract entity type from GID
CREATE OR REPLACE FUNCTION extract_entity_type(gid bytea)
RETURNS int AS $func$
BEGIN
    RETURN get_byte(gid, 8) * 256 + get_byte(gid, 9);
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

-- Extract timestamp from GID
CREATE OR REPLACE FUNCTION extract_timestamp(gid bytea)
RETURNS timestamp AS $func$
DECLARE
    ms bigint;
BEGIN
    ms := (get_byte(gid, 10)::bigint << 56) |
          (get_byte(gid, 11)::bigint << 48) |
          (get_byte(gid, 12)::bigint << 40) |
          (get_byte(gid, 13)::bigint << 32) |
          (get_byte(gid, 14)::bigint << 24) |
          (get_byte(gid, 15)::bigint << 16) |
          (get_byte(gid, 16)::bigint << 8) |
          get_byte(gid, 17)::bigint;
          
    RETURN to_timestamp(ms / 1000.0);
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

-- Example usage:
-- Generate a new GID for entity type 42 with a specific tenant ID
-- SELECT gid_to_string(generate_gid(generate_tenant_id(), 42));

-- Parse a GID from string
-- SELECT parse_gid('your-base64-encoded-gid-here');

-- Extract components
-- SELECT 
--   gid_to_string(gid) as gid_string,
--   tenant_id_to_string(extract_tenant_id(gid)) as tenant_id,
--   extract_entity_type(gid) as entity_type,
--   extract_timestamp(gid) as created_at
-- FROM (SELECT generate_gid(generate_tenant_id(), 42) as gid) t;

CREATE OR REPLACE FUNCTION decode_base64_unpadded(input_text text) 
RETURNS bytea AS $$
DECLARE
    padded_text text;
    mod_length integer;
    normalized_text text;
BEGIN
    -- Replace URL-safe characters with standard base64 characters
    normalized_text := translate(input_text, '-_', '+/');
    
    -- Calculate how many padding characters we need to add
    mod_length := length(normalized_text) % 4;
    
    -- Add the required padding
    IF mod_length = 0 THEN
        padded_text := normalized_text;
    ELSIF mod_length = 1 THEN
        -- Invalid base64 - length mod 4 can't be 1
        RAISE EXCEPTION 'Invalid base64 length';
    ELSIF mod_length = 2 THEN
        padded_text := normalized_text || '==';
    ELSIF mod_length = 3 THEN
        padded_text := normalized_text || '=';
    END IF;
    
    -- Decode the padded base64
    RETURN decode(padded_text, 'base64');
END;
$$ LANGUAGE plpgsql;

CREATE TYPE data_sensitivity AS ENUM ('NONE', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE business_impact AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');

CREATE TABLE risk_assessments (
    tenant_id TEXT NOT NULL,
    id TEXT PRIMARY KEY,
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    accessed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accessed_by TEXT NOT NULL REFERENCES peoples(id) ON DELETE SET NULL,
    approved_at TIMESTAMP WITH TIME ZONE NOT NULL,
    approved_by TEXT NOT NULL REFERENCES peoples(id) ON DELETE SET NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    data_sensitivity data_sensitivity NOT NULL,
    business_impact business_impact NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

ALTER TABLE peoples ADD COLUMN user_id TEXT REFERENCES users(id) ON DELETE SET NULL;

UPDATE peoples p
SET user_id = u.id 
FROM users u
WHERE u.email_address = p.primary_email_address;

INSERT INTO peoples (
    tenant_id,
    id,
    organization_id,
    full_name,
    primary_email_address,
    kind,
    user_id,
    additional_email_addresses,
    created_at,
    updated_at
)
SELECT 
    o.tenant_id,
    generate_gid(decode_base64_unpadded(o.tenant_id), 8),
    o.id,
    u.fullname,
    u.email_address,
    'CONTRACTOR',
    u.id,
    ARRAY[]::TEXT[],
    NOW(),
    NOW()
FROM users u
JOIN users_organizations uo ON u.id = uo.user_id
JOIN organizations o ON uo.organization_id = o.id
LEFT JOIN peoples p ON u.email_address = p.primary_email_address
WHERE p.id IS NULL;

