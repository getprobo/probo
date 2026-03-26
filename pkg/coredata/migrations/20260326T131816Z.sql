CREATE TYPE evidence_description_status AS ENUM ('PENDING', 'PROCESSING', 'COMPLETED');

ALTER TABLE evidences
    ADD COLUMN description_status evidence_description_status NOT NULL DEFAULT 'PENDING',
    ADD COLUMN description_processing_started_at TIMESTAMPTZ;

UPDATE evidences SET description_status = 'COMPLETED' WHERE evidence_file_id IS NULL;

ALTER TABLE evidences
    ALTER COLUMN description_status DROP DEFAULT;
