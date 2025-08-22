ALTER TABLE trust_center_accesses ADD COLUMN active BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE trust_center_accesses ALTER COLUMN active DROP DEFAULT;
