-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
