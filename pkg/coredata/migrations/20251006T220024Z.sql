-- Create authorization tables for the new authz service
-- This migration creates the new authz tables while keeping the existing users_organizations table
-- for backward compatibility during the transition

-- Create role enum
CREATE TYPE authz_role AS ENUM ('OWNER', 'ADMIN', 'MEMBER', 'VIEWER');

-- Create authz_memberships table with id as primary key
CREATE TABLE authz_memberships (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    role authz_role NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (user_id, organization_id)
);

-- Create authz_invitations table
CREATE TABLE authz_invitations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    email TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role authz_role NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

-- Copy data from users_organizations to authz_memberships
INSERT INTO authz_memberships (tenant_id, id, user_id, organization_id, role, created_at, updated_at)
SELECT
    organizations.tenant_id,
    generate_gid(decode_base64_unpadded(organizations.tenant_id), 39) as id,
    users_organizations.user_id,
    users_organizations.organization_id,
    'MEMBER'::authz_role as role,  -- Default role for existing memberships
    users_organizations.created_at,
    users_organizations.created_at as updated_at
FROM users_organizations
JOIN organizations ON users_organizations.organization_id = organizations.id;
