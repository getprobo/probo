-- Rename priority to rank
ALTER TABLE tasks RENAME COLUMN priority TO rank;

ALTER TABLE tasks DROP CONSTRAINT tasks_organization_id_state_priority_key;

-- Add task priority enum
CREATE TYPE task_priority AS ENUM ('URGENT', 'HIGH', 'MEDIUM', 'LOW');

ALTER TABLE tasks ADD COLUMN priority task_priority NOT NULL DEFAULT 'MEDIUM'::task_priority;

ALTER TABLE tasks ALTER COLUMN priority DROP DEFAULT;

-- Rank is now scoped to (state, priority) — backfill ranks per group
WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (
        PARTITION BY organization_id, state, priority
        ORDER BY rank
    ) AS new_rank
    FROM tasks
)
UPDATE tasks SET rank = ranked.new_rank FROM ranked WHERE tasks.id = ranked.id;

ALTER TABLE tasks
ADD CONSTRAINT tasks_organization_id_state_priority_rank_key
    UNIQUE (organization_id, state, priority, rank)
    DEFERRABLE INITIALLY DEFERRED;

-- Computed column for composite ordering (priority level then rank)
ALTER TABLE tasks ADD COLUMN priority_rank int GENERATED ALWAYS AS (
    (CASE priority
        WHEN 'URGENT' THEN 1
        WHEN 'HIGH' THEN 2
        WHEN 'MEDIUM' THEN 3
        WHEN 'LOW' THEN 4
    END) * 1000000 + rank
) STORED;
