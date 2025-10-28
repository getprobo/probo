ALTER TABLE trust_center_references ADD COLUMN rank INTEGER;

WITH ranked_references AS (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY trust_center_id ORDER BY created_at DESC, id DESC) AS rn
    FROM trust_center_references
)
UPDATE trust_center_references tcr
SET rank = rr.rn
FROM ranked_references rr
WHERE tcr.id = rr.id;

ALTER TABLE trust_center_references ALTER COLUMN rank SET NOT NULL;

ALTER TABLE trust_center_references
ADD CONSTRAINT trust_center_references_trust_center_id_rank_key
    UNIQUE (trust_center_id, rank)
    DEFERRABLE INITIALLY DEFERRED;
