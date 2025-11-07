-- Set all existing memberships to OWNER role
UPDATE memberships SET role = 'OWNER';

ALTER TABLE controls ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE controls
SET organization_id = organizations.id
FROM organizations
WHERE controls.tenant_id = organizations.tenant_id
AND controls.organization_id IS NULL;
ALTER TABLE controls ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE evidences ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE evidences
SET organization_id = organizations.id
FROM organizations
WHERE evidences.tenant_id = organizations.tenant_id
AND evidences.organization_id IS NULL;
ALTER TABLE evidences ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE files ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE files
SET organization_id = organizations.id
FROM organizations
WHERE files.tenant_id = organizations.tenant_id
AND files.organization_id IS NULL;
ALTER TABLE files ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE document_versions ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE document_versions
SET organization_id = organizations.id
FROM organizations
WHERE document_versions.tenant_id = organizations.tenant_id
AND document_versions.organization_id IS NULL;
ALTER TABLE document_versions ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE document_version_signatures ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE document_version_signatures
SET organization_id = organizations.id
FROM organizations
WHERE document_version_signatures.tenant_id = organizations.tenant_id
AND document_version_signatures.organization_id IS NULL;
ALTER TABLE document_version_signatures ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE trust_center_references ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE trust_center_references
SET organization_id = organizations.id
FROM organizations
WHERE trust_center_references.tenant_id = organizations.tenant_id
AND trust_center_references.organization_id IS NULL;
ALTER TABLE trust_center_references ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE trust_center_accesses ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE trust_center_accesses
SET organization_id = organizations.id
FROM organizations
WHERE trust_center_accesses.tenant_id = organizations.tenant_id
AND trust_center_accesses.organization_id IS NULL;
ALTER TABLE trust_center_accesses ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE trust_center_document_accesses ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE trust_center_document_accesses
SET organization_id = organizations.id
FROM organizations
WHERE trust_center_document_accesses.tenant_id = organizations.tenant_id
AND trust_center_document_accesses.organization_id IS NULL;
ALTER TABLE trust_center_document_accesses ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_trust_center_document_accesses_organization_id ON trust_center_document_accesses(organization_id);

ALTER TABLE vendor_services ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE vendor_services
SET organization_id = organizations.id
FROM organizations
WHERE vendor_services.tenant_id = organizations.tenant_id
AND vendor_services.organization_id IS NULL;
ALTER TABLE vendor_services ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_vendor_services_organization_id ON vendor_services(organization_id);

ALTER TABLE vendor_contacts ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE vendor_contacts
SET organization_id = organizations.id
FROM organizations
WHERE vendor_contacts.tenant_id = organizations.tenant_id
AND vendor_contacts.organization_id IS NULL;
ALTER TABLE vendor_contacts ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_vendor_contacts_organization_id ON vendor_contacts(organization_id);

ALTER TABLE vendor_risk_assessments ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE vendor_risk_assessments
SET organization_id = organizations.id
FROM organizations
WHERE vendor_risk_assessments.tenant_id = organizations.tenant_id
AND vendor_risk_assessments.organization_id IS NULL;
ALTER TABLE vendor_risk_assessments ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_vendor_risk_assessments_organization_id ON vendor_risk_assessments(organization_id);

ALTER TABLE reports ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE reports
SET organization_id = organizations.id
FROM organizations
WHERE reports.tenant_id = organizations.tenant_id
AND reports.organization_id IS NULL;
ALTER TABLE reports ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reports_organization_id ON reports(organization_id);

ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE custom_domains
SET organization_id = organizations.id
FROM organizations
WHERE custom_domains.tenant_id = organizations.tenant_id
AND custom_domains.organization_id IS NULL;
ALTER TABLE custom_domains ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_custom_domains_organization_id ON custom_domains(organization_id);
