CREATE TYPE session_auth_method AS ENUM (
    'PASSWORD',
    'SAML'
);

ALTER TABLE sessions ADD COLUMN auth_method session_auth_method;
ALTER TABLE sessions ADD COLUMN authenticated_at TIMESTAMP;
ALTER TABLE sessions ADD COLUMN membership_id TEXT REFERENCES authz_memberships(id) ON DELETE CASCADE;

-- Expire all existing sessions as they are not backward compatible
UPDATE sessions SET 
    expire_reason = 'revoked',
    expired_at = NOW(),
    updated_at = NOW(),
    auth_method = 'PASSWORD',
    authenticated_at = created_at
WHERE expire_reason IS NULL;

UPDATE sessions SET 
    auth_method = 'PASSWORD',
    authenticated_at = created_at
WHERE auth_method IS NULL;

ALTER TABLE sessions ALTER COLUMN auth_method SET NOT NULL;
ALTER TABLE sessions ALTER COLUMN authenticated_at SET NOT NULL;
