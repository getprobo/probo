ALTER TABLE nonconformities DROP CONSTRAINT IF EXISTS nonconformities_audit_id_fkey;
ALTER TABLE nonconformities ALTER COLUMN audit_id DROP NOT NULL;
ALTER TABLE nonconformities ADD CONSTRAINT nonconformities_audit_id_fkey
    FOREIGN KEY (audit_id) REFERENCES audits(id) ON UPDATE CASCADE ON DELETE SET NULL;
