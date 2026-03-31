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

-- OAuth2 Authorization Server tables

CREATE TYPE oauth2_client_visibility AS ENUM ('private', 'public');
CREATE TYPE oauth2_client_token_endpoint_auth_method AS ENUM ('client_secret_basic', 'client_secret_post', 'none');
CREATE TYPE oauth2_device_code_status AS ENUM ('pending', 'authorized', 'denied', 'expired');

CREATE TABLE iam_oauth2_clients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    client_secret_hash BYTEA,
    client_name TEXT NOT NULL,
    visibility oauth2_client_visibility NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    scopes TEXT[] NOT NULL,
    grant_types TEXT[] NOT NULL,
    response_types TEXT[] NOT NULL,
    token_endpoint_auth_method oauth2_client_token_endpoint_auth_method NOT NULL,
    logo_uri TEXT,
    client_uri TEXT,
    contacts TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE iam_oauth2_authorization_codes (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL REFERENCES iam_oauth2_clients(id),
    identity_id TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    code_challenge TEXT,
    code_challenge_method TEXT,
    nonce TEXT,
    auth_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE iam_oauth2_access_tokens (
    id TEXT PRIMARY KEY,
    hashed_value BYTEA NOT NULL,
    client_id TEXT NOT NULL REFERENCES iam_oauth2_clients(id),
    identity_id TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT iam_oauth2_access_tokens_hashed_value_unique UNIQUE (hashed_value)
);

CREATE TABLE iam_oauth2_refresh_tokens (
    id TEXT PRIMARY KEY,
    hashed_value BYTEA NOT NULL,
    client_id TEXT NOT NULL REFERENCES iam_oauth2_clients(id),
    identity_id TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    access_token_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT iam_oauth2_refresh_tokens_hashed_value_unique UNIQUE (hashed_value)
);

CREATE TABLE iam_oauth2_device_codes (
    id TEXT PRIMARY KEY,
    device_code_hash BYTEA NOT NULL,
    user_code TEXT NOT NULL,
    client_id TEXT NOT NULL REFERENCES iam_oauth2_clients(id),
    scopes TEXT[] NOT NULL,
    identity_id TEXT,
    status oauth2_device_code_status NOT NULL,
    last_polled_at TIMESTAMP WITH TIME ZONE,
    poll_interval INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT iam_oauth2_device_codes_device_code_hash_unique UNIQUE (device_code_hash),
    CONSTRAINT iam_oauth2_device_codes_user_code_unique UNIQUE (user_code)
);

CREATE TABLE iam_oauth2_consents (
    id TEXT PRIMARY KEY,
    identity_id TEXT NOT NULL,
    client_id TEXT NOT NULL REFERENCES iam_oauth2_clients(id),
    scopes TEXT[] NOT NULL,
    redirect_uri TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    code_challenge_method TEXT NOT NULL,
    nonce TEXT NOT NULL,
    state TEXT NOT NULL,
    approved BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
