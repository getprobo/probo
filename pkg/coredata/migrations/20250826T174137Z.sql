CREATE TABLE controls_snapshots (
    control_id TEXT NOT NULL REFERENCES controls(id) ON DELETE CASCADE ON UPDATE CASCADE,
    snapshot_id TEXT NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE ON UPDATE CASCADE,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (control_id, snapshot_id)
);

ALTER TABLE controls_audits DROP CONSTRAINT controls_audits_control_id_fkey;
ALTER TABLE controls_audits ADD CONSTRAINT controls_audits_control_id_fkey
    FOREIGN KEY (control_id) REFERENCES controls(id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE controls_audits DROP CONSTRAINT controls_audits_audit_id_fkey;
ALTER TABLE controls_audits ADD CONSTRAINT controls_audits_audit_id_fkey
    FOREIGN KEY (audit_id) REFERENCES audits(id) ON DELETE CASCADE ON UPDATE CASCADE;
