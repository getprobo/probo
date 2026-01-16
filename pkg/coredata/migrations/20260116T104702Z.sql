CREATE TABLE iam_tokens (
    id TEXT PRIMARY KEY,
    hashed_value BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT iam_tokens_hashed_value_unique UNIQUE (hashed_value)
);
