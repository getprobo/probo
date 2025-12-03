CREATE TYPE session_expire_reason AS ENUM (
    'idle_timeout',
    'revoked',
    'closed'
);

ALTER TABLE sessions ADD COLUMN expire_reason session_expire_reason;

UPDATE sessions SET expire_reason = 'idle_timeout' WHERE expired_at < NOW();
