ALTER TYPE session_auth_method ADD VALUE 'GOOGLE';
ALTER TYPE session_auth_method ADD VALUE 'MICROSOFT';

CREATE TYPE iam_oidc_provider AS ENUM (
    'GOOGLE',
    'MICROSOFT'
);

CREATE TABLE iam_oidc_states (
    id TEXT PRIMARY KEY,
    provider iam_oidc_provider NOT NULL,
    nonce TEXT NOT NULL,
    code_verifier TEXT NOT NULL,
    continue_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_iam_oidc_states_expires_at ON iam_oidc_states(expires_at);
