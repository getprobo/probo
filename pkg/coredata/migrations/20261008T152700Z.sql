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

CREATE TABLE internal_controls_tasks (
    internal_control_id TEXT NOT NULL REFERENCES internal_controls(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (internal_control_id, task_id),
    CONSTRAINT internal_controls_tasks_internal_control_reference_id_key UNIQUE (internal_control_id, reference_id)
);

INSERT INTO internal_controls_tasks (
    internal_control_id,
    task_id,
    organization_id,
    tenant_id,
    reference_id,
    created_at
)
SELECT
    internal_control_id,
    id,
    organization_id,
    tenant_id,
    reference_id,
    created_at
FROM
    tasks
WHERE
    internal_control_id IS NOT NULL;

ALTER TABLE tasks DROP CONSTRAINT tasks_reference_id_unique;

ALTER TABLE tasks DROP CONSTRAINT fk_tasks_mitigation_id;
ALTER TABLE tasks
    ADD CONSTRAINT tasks_internal_control_id_fkey
    FOREIGN KEY (internal_control_id) REFERENCES internal_controls(id) ON DELETE SET NULL;

-- TODO: drop tasks.internal_control_id in a later commit, once nothing reads it.
