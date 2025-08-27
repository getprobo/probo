-- Create audit_reports junction table to support multiple file attachments per audit
CREATE TABLE audit_reports (
    audit_id TEXT NOT NULL REFERENCES audits(id) ON DELETE CASCADE ON UPDATE CASCADE,
    report_id TEXT NOT NULL REFERENCES reports(id) ON DELETE CASCADE ON UPDATE CASCADE,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (audit_id, report_id)
);

-- Migrate existing audit report relationships to the new junction table
INSERT INTO audit_reports (audit_id, report_id, tenant_id, created_at)
SELECT id, report_id, organization_id, updated_at
FROM audits 
WHERE report_id IS NOT NULL;

-- Remove the report_id column from audits table (will be handled via junction table)
ALTER TABLE audits DROP COLUMN report_id;