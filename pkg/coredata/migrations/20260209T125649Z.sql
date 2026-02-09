DELETE FROM processing_activity_data_protection_impact_assessments
WHERE id IN (
    SELECT id
    FROM (
        SELECT id,
               ROW_NUMBER() OVER (
                   PARTITION BY processing_activity_id, COALESCE(snapshot_id, '')
                   ORDER BY updated_at DESC
               ) AS rn
        FROM processing_activity_data_protection_impact_assessments
    ) t
    WHERE rn > 1
);

DELETE FROM processing_activity_transfer_impact_assessments
WHERE id IN (
    SELECT id
    FROM (
        SELECT id,
               ROW_NUMBER() OVER (
                   PARTITION BY processing_activity_id, COALESCE(snapshot_id, '')
                   ORDER BY updated_at DESC
               ) AS rn
        FROM processing_activity_transfer_impact_assessments
    ) t
    WHERE rn > 1
);

CREATE UNIQUE INDEX processing_activity_dpias_processing_activity_id_snapshot_id_uniq
    ON processing_activity_data_protection_impact_assessments (processing_activity_id, COALESCE(snapshot_id, ''));

CREATE UNIQUE INDEX processing_activity_tias_processing_activity_id_snapshot_id_uniq
    ON processing_activity_transfer_impact_assessments (processing_activity_id, COALESCE(snapshot_id, ''));
