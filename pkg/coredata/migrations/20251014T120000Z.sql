-- Add retry tracking fields for SSL certificate provisioning
ALTER TABLE custom_domains ADD COLUMN ssl_retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE custom_domains ADD COLUMN ssl_last_attempt_at TIMESTAMP;

-- Add index to efficiently find stale provisioning attempts
CREATE INDEX idx_custom_domains_ssl_last_attempt
    ON custom_domains(ssl_last_attempt_at, ssl_status)
    WHERE ssl_status IN ('PENDING', 'PROVISIONING', 'RENEWING');
