-- Set all existing memberships to OWNER role
UPDATE authz_memberships SET role = 'OWNER';

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

ALTER TABLE vendor_services ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE vendor_services
SET organization_id = organizations.id
FROM organizations
WHERE vendor_services.tenant_id = organizations.tenant_id
AND vendor_services.organization_id IS NULL;
ALTER TABLE vendor_services ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE vendor_contacts ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE vendor_contacts
SET organization_id = organizations.id
FROM organizations
WHERE vendor_contacts.tenant_id = organizations.tenant_id
AND vendor_contacts.organization_id IS NULL;
ALTER TABLE vendor_contacts ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE vendor_risk_assessments ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE vendor_risk_assessments
SET organization_id = organizations.id
FROM organizations
WHERE vendor_risk_assessments.tenant_id = organizations.tenant_id
AND vendor_risk_assessments.organization_id IS NULL;
ALTER TABLE vendor_risk_assessments ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE reports ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE reports
SET organization_id = organizations.id
FROM organizations
WHERE reports.tenant_id = organizations.tenant_id
AND reports.organization_id IS NULL;
ALTER TABLE reports ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE custom_domains
SET organization_id = organizations.id
FROM organizations
WHERE custom_domains.tenant_id = organizations.tenant_id
AND custom_domains.organization_id IS NULL;
ALTER TABLE custom_domains ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE asset_vendors ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE asset_vendors
SET organization_id = organizations.id
FROM organizations
WHERE asset_vendors.tenant_id = organizations.tenant_id
AND asset_vendors.organization_id IS NULL;
ALTER TABLE asset_vendors ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE authz_api_keys_memberships ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE authz_api_keys_memberships
SET organization_id = organizations.id
FROM organizations
WHERE authz_api_keys_memberships.tenant_id = organizations.tenant_id
AND authz_api_keys_memberships.organization_id IS NULL;
ALTER TABLE authz_api_keys_memberships ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE controls_audits ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE controls_audits
SET organization_id = organizations.id
FROM organizations
WHERE controls_audits.tenant_id = organizations.tenant_id
AND controls_audits.organization_id IS NULL;
ALTER TABLE controls_audits ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE controls_documents ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE controls_documents
SET organization_id = organizations.id
FROM organizations
WHERE controls_documents.tenant_id = organizations.tenant_id
AND controls_documents.organization_id IS NULL;
ALTER TABLE controls_documents ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE controls_measures ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE controls_measures
SET organization_id = organizations.id
FROM organizations
WHERE controls_measures.tenant_id = organizations.tenant_id
AND controls_measures.organization_id IS NULL;
ALTER TABLE controls_measures ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE controls_snapshots ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE controls_snapshots
SET organization_id = organizations.id
FROM organizations
WHERE controls_snapshots.tenant_id = organizations.tenant_id
AND controls_snapshots.organization_id IS NULL;
ALTER TABLE controls_snapshots ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE data_vendors ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE data_vendors
SET organization_id = organizations.id
FROM organizations
WHERE data_vendors.tenant_id = organizations.tenant_id
AND data_vendors.organization_id IS NULL;
ALTER TABLE data_vendors ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE export_jobs ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE export_jobs
SET organization_id = organizations.id
FROM organizations
WHERE export_jobs.tenant_id = organizations.tenant_id
AND export_jobs.organization_id IS NULL;
ALTER TABLE export_jobs ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE processing_activity_vendors ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE processing_activity_vendors
SET organization_id = organizations.id
FROM organizations
WHERE processing_activity_vendors.tenant_id = organizations.tenant_id
AND processing_activity_vendors.organization_id IS NULL;
ALTER TABLE processing_activity_vendors ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE risks_documents ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE risks_documents
SET organization_id = organizations.id
FROM organizations
WHERE risks_documents.tenant_id = organizations.tenant_id
AND risks_documents.organization_id IS NULL;
ALTER TABLE risks_documents ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE risks_measures ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE risks_measures
SET organization_id = organizations.id
FROM organizations
WHERE risks_measures.tenant_id = organizations.tenant_id
AND risks_measures.organization_id IS NULL;
ALTER TABLE risks_measures ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE risks_obligations ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE risks_obligations
SET organization_id = organizations.id
FROM organizations
WHERE risks_obligations.tenant_id = organizations.tenant_id
AND risks_obligations.organization_id IS NULL;
ALTER TABLE risks_obligations ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE vendor_compliance_reports ADD COLUMN IF NOT EXISTS organization_id TEXT;
UPDATE vendor_compliance_reports
SET organization_id = organizations.id
FROM organizations
WHERE vendor_compliance_reports.tenant_id = organizations.tenant_id
AND vendor_compliance_reports.organization_id IS NULL;
ALTER TABLE vendor_compliance_reports ALTER COLUMN organization_id SET NOT NULL;
