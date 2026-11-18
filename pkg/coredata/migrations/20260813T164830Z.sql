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

ALTER TABLE agent_runs RENAME TO agent_executions;

ALTER TABLE agent_executions
    RENAME CONSTRAINT agent_runs_pkey TO agent_executions_pkey;

ALTER TABLE agent_executions
    RENAME CONSTRAINT agent_runs_organization_id_fkey TO agent_executions_organization_id_fkey;

ALTER INDEX agent_runs_source_session_key
    RENAME TO agent_executions_source_session_key;

ALTER TABLE agent_inputs
    RENAME COLUMN agent_run_id TO agent_execution_id;

ALTER TABLE agent_inputs
    RENAME CONSTRAINT agent_inputs_agent_run_id_fkey TO agent_inputs_agent_execution_id_fkey;

ALTER TABLE agent_execution_anchors
    RENAME COLUMN agent_run_id TO agent_execution_id;

ALTER TABLE agent_execution_anchors
    RENAME CONSTRAINT agent_execution_anchors_agent_run_id_fkey
    TO agent_execution_anchors_agent_execution_id_fkey;
