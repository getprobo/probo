CREATE TYPE trust_center_access_state AS ENUM ('ACTIVE', 'INACTIVE');

ALTER TABLE trust_center_accesses
  ADD COLUMN state trust_center_access_state NOT NULL DEFAULT 'INACTIVE';

UPDATE trust_center_accesses
  SET state = 'ACTIVE'
  WHERE active = true;

ALTER TABLE trust_center_accesses
  DROP COLUMN active;

ALTER TABLE trust_center_accesses
  ALTER COLUMN state DROP DEFAULT;