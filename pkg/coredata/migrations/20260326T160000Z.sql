ALTER TABLE tasks ADD COLUMN priority INTEGER;

WITH ranked_tasks AS (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY organization_id, state ORDER BY created_at DESC, id DESC) AS rn
    FROM tasks
)
UPDATE tasks t
SET priority = rt.rn
FROM ranked_tasks rt
WHERE t.id = rt.id;

ALTER TABLE tasks ALTER COLUMN priority SET NOT NULL;

ALTER TABLE tasks
ADD CONSTRAINT tasks_organization_id_state_priority_key
    UNIQUE (organization_id, state, priority)
    DEFERRABLE INITIALLY DEFERRED;
