ALTER TABLE users RENAME TO identities;
ALTER TABLE sessions RENAME COLUMN user_id TO identity_id;
ALTER TABLE authz_memberships RENAME COLUMN user_id TO identity_id;
ALTER TABLE peoples RENAME COLUMN user_id TO identity_id;
ALTER TABLE auth_user_api_keys RENAME TO auth_personal_api_keys;
ALTER TABLE auth_personal_api_keys RENAME COLUMN user_id TO identity_id;
ALTER TABLE authz_api_keys_memberships RENAME COLUMN auth_user_api_key_id TO auth_personal_api_key_id;

