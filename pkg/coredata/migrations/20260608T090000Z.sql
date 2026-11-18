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

-- Drop the agent-run worker lease. Recovery from a crashed worker is now
-- manual (inspect logs, move the row out of RUNNING by hand). Known stops
-- transition explicitly: graceful shutdown returns the row to PENDING and
-- approval pauses park it in AWAITING_APPROVAL, so nothing relies on a
-- timeout to requeue. Re-introduce leasing here when the system matures.

DROP INDEX IF EXISTS idx_agent_runs_running_lease;

ALTER TABLE agent_runs
	DROP COLUMN IF EXISTS lease_owner,
	DROP COLUMN IF EXISTS lease_expires_at,
	DROP COLUMN IF EXISTS lease_generation;
