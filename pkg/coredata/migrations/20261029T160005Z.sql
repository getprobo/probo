-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

ALTER TABLE audit_log_entries
    ADD COLUMN outcome TEXT NOT NULL DEFAULT 'ALLOWED',
    ADD COLUMN denial_reason TEXT,
    ADD COLUMN actor_email TEXT,
    ADD COLUMN actor_name TEXT,
    ADD COLUMN origin_ip TEXT,
    ADD COLUMN user_agent TEXT,
    ADD COLUMN request_id TEXT,
    ADD COLUMN recorded_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp();

UPDATE audit_log_entries
SET recorded_at = created_at;

ALTER TABLE audit_log_entries
    ALTER COLUMN outcome DROP DEFAULT,
    ALTER COLUMN recorded_at DROP DEFAULT;

ALTER TABLE audit_log_entries
    ADD CONSTRAINT audit_log_entries_outcome_check
    CHECK (outcome IN ('ALLOWED', 'DENIED'));
