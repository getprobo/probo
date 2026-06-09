ALTER TABLE devices
    ADD COLUMN enrollment_token_id TEXT REFERENCES device_enrollment_tokens(id) ON UPDATE CASCADE ON DELETE SET NULL;

CREATE INDEX devices_enrollment_token_id_enrolled_at_idx
    ON devices (enrollment_token_id, enrolled_at DESC)
    WHERE enrollment_token_id IS NOT NULL;
