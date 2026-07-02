-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

DROP TABLE control_state_transitions;
DROP TABLE evidence_state_transitions;
DROP TABLE task_state_transitions;

ALTER TABLE tasks ADD COLUMN control_id TEXT;

UPDATE tasks
SET control_id = controls_tasks.control_id
FROM controls_tasks
WHERE tasks.id = controls_tasks.task_id;

ALTER TABLE tasks ALTER COLUMN control_id SET NOT NULL;

ALTER TABLE tasks ADD CONSTRAINT fk_tasks_control_id
    FOREIGN KEY (control_id) REFERENCES controls(id) ON DELETE CASCADE;

DROP TABLE controls_tasks;
