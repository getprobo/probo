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
