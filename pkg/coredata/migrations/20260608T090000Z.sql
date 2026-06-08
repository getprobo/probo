-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

-- Drop the agent-run worker lease. Recovery from a crashed worker is now
-- manual (inspect logs, move the row out of RUNNING by hand). Known stops
-- transition explicitly: graceful shutdown returns the row to PENDING and
-- approval pauses park it in AWAITING_APPROVAL, so nothing relies on a
-- timeout to requeue. Re-introduce leasing here when the system matures.

DROP INDEX IF EXISTS idx_agent_runs_running_lease;

ALTER TABLE agent_runs
	DROP COLUMN lease_expires_at,
	DROP COLUMN lease_generation;
