-- Create authorization tables for the new authz service
-- This migration creates the new authz tables while keeping the existing users_organizations table
-- for backward compatibility during the transition

-- Create role enum
CREATE TYPE authz_role AS ENUM ('owner', 'admin', 'member', 'viewer');

-- Create authz_memberships table (replaces users_organizations)
CREATE TABLE authz_memberships (
    user_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    role authz_role NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, organization_id)
);

-- Create authz_invitations table
CREATE TABLE authz_invitations (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    email TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role authz_role NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

-- Copy data from users_organizations to authz_memberships
INSERT INTO authz_memberships (user_id, organization_id, role, created_at, updated_at)
SELECT 
    user_id, 
    organization_id, 
    'member'::authz_role as role,  -- Default role for existing memberships
    created_at,
    created_at as updated_at
FROM users_organizations;