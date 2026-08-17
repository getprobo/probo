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

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM agent_executions
        WHERE status = 'RUNNING'
            AND processing_owner_token IS NOT NULL
    ) THEN
        RAISE EXCEPTION
            'cannot remap RUNNING agent executions that still hold a processing lease';
    END IF;
END $$;

UPDATE agent_executions
SET status = 'IDLE',
    started_at = NULL
WHERE status IN ('PENDING', 'RUNNING')
    AND processing_owner_token IS NULL;

ALTER TABLE agent_executions
    ALTER COLUMN status SET DEFAULT 'IDLE';

ALTER TABLE agent_executions
    ALTER COLUMN status DROP DEFAULT;
