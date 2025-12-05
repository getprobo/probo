CREATE TYPE trust_center_document_access_status AS ENUM ('REQUESTED', 'GRANTED', 'REJECTED', 'REVOKED');

ALTER TABLE
    trust_center_document_accesses
ADD
    COLUMN status trust_center_document_access_status;

UPDATE
    trust_center_document_accesses
SET
    status = CASE
        WHEN active = 'T' THEN 'GRANTED' :: trust_center_document_access_status
        ELSE 'REQUESTED' :: trust_center_document_access_status
    END;

ALTER TABLE
    trust_center_document_accesses
ALTER COLUMN
    status
SET
    NOT NULL;