ALTER TABLE emails
ADD COLUMN status TEXT NOT NULL DEFAULT 'PENDING',
ADD COLUMN processing_started_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN attempt_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 10,
ADD COLUMN last_attempted_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN last_error TEXT;

UPDATE emails SET status = 'SENT' WHERE sent_at IS NOT NULL;

CREATE INDEX idx_emails_pending ON emails (status, created_at)
WHERE status = 'PENDING';
