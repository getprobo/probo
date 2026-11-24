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

-- Claim walks QUEUED rows by (created_at, id). The same-source anti-join
-- looks up RUNNING / older QUEUED by (access_review_source_id, status,
-- created_at, id). Stale recovery walks RUNNING by updated_at. Partial +
-- composite indexes keep these paths from rescanning terminal history.

CREATE INDEX idx_access_source_sync_runs_queued
    ON access_source_sync_runs (created_at, id)
    WHERE status = 'QUEUED';

CREATE INDEX idx_access_source_sync_runs_running
    ON access_source_sync_runs (updated_at, id)
    WHERE status = 'RUNNING';

CREATE INDEX idx_access_source_sync_runs_source_status
    ON access_source_sync_runs (access_review_source_id, status, created_at, id);
