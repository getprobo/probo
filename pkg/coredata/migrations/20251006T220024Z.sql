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
