// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package probo

// Probo Service Actions
// Format: core:<entity>:<action>
const (
	// Organization actions
	ActionOrganizationGet                  = "core:organization:get"
	ActionOrganizationGetLogoUrl           = "core:organization:get-logo-url"
	ActionOrganizationGetHorizontalLogoUrl = "core:organization:get-horizontal-logo-url"

	// OrganizationContext actions
	ActionOrganizationContextGet    = "core:organization-context:get"
	ActionOrganizationContextUpdate = "core:organization-context:update"

	// TrustCenter actions
	ActionTrustCenterGet                          = "core:trust-center:get"
	ActionTrustCenterUpdate                       = "core:trust-center:update"
	ActionTrustCenterGetNda                       = "core:trust-center:get-nda"
	ActionTrustCenterNonDisclosureAgreementUpload = "core:trust-center:upload-nda"
	ActionTrustCenterNonDisclosureAgreementDelete = "core:trust-center:delete-nda"

	// TrustCenterAccess actions
	ActionTrustCenterAccessGet    = "core:trust-center-access:get"
	ActionTrustCenterAccessList   = "core:trust-center-access:list"
	ActionTrustCenterAccessCreate = "core:trust-center-access:create"
	ActionTrustCenterAccessUpdate = "core:trust-center-access:update"
	ActionTrustCenterAccessDelete = "core:trust-center-access:delete"

	// TrustCenterReference actions
	ActionTrustCenterReferenceList       = "core:trust-center-reference:list"
	ActionTrustCenterReferenceGetLogoUrl = "core:trust-center-reference:get-logo-url"
	ActionTrustCenterReferenceCreate     = "core:trust-center-reference:create"
	ActionTrustCenterReferenceUpdate     = "core:trust-center-reference:update"
	ActionTrustCenterReferenceDelete     = "core:trust-center-reference:delete"

	// TrustCenterFile actions
	ActionTrustCenterFileGet        = "core:trust-center-file:get"
	ActionTrustCenterFileList       = "core:trust-center-file:list"
	ActionTrustCenterFileGetFileUrl = "core:trust-center-file:get-file-url"
	ActionTrustCenterFileUpdate     = "core:trust-center-file:update"
	ActionTrustCenterFileDelete     = "core:trust-center-file:delete"
	ActionTrustCenterFileCreate     = "core:trust-center-file:create"

	// Vendor actions
	ActionVendorList   = "core:vendor:list"
	ActionVendorGet    = "core:vendor:get"
	ActionVendorCreate = "core:vendor:create"
	ActionVendorUpdate = "core:vendor:update"
	ActionVendorDelete = "core:vendor:delete"
	ActionVendorAssess = "core:vendor:assess"

	// VendorContact actions
	ActionVendorContactGet    = "core:vendor-contact:get"
	ActionVendorContactList   = "core:vendor-contact:list"
	ActionVendorContactCreate = "core:vendor-contact:create"
	ActionVendorContactUpdate = "core:vendor-contact:update"
	ActionVendorContactDelete = "core:vendor-contact:delete"

	// VendorService actions
	ActionVendorServiceGet    = "core:vendor-service:get"
	ActionVendorServiceList   = "core:vendor-service:list"
	ActionVendorServiceCreate = "core:vendor-service:create"
	ActionVendorServiceUpdate = "core:vendor-service:update"
	ActionVendorServiceDelete = "core:vendor-service:delete"

	// VendorComplianceReport actions
	ActionVendorComplianceReportGet    = "core:vendor-compliance-report:get"
	ActionVendorComplianceReportList   = "core:vendor-compliance-report:list"
	ActionVendorComplianceReportUpload = "core:vendor-compliance-report:upload"
	ActionVendorComplianceReportDelete = "core:vendor-compliance-report:delete"

	// VendorBusinessAssociateAgreement actions
	ActionVendorBusinessAssociateAgreementGet    = "core:vendor-business-associate-agreement:get"
	ActionVendorBusinessAssociateAgreementUpload = "core:vendor-business-associate-agreement:upload"
	ActionVendorBusinessAssociateAgreementUpdate = "core:vendor-business-associate-agreement:update"
	ActionVendorBusinessAssociateAgreementDelete = "core:vendor-business-associate-agreement:delete"

	// VendorDataPrivacyAgreement actions
	ActionVendorDataPrivacyAgreementGet    = "core:vendor-data-privacy-agreement:get"
	ActionVendorDataPrivacyAgreementUpload = "core:vendor-data-privacy-agreement:upload"
	ActionVendorDataPrivacyAgreementUpdate = "core:vendor-data-privacy-agreement:update"
	ActionVendorDataPrivacyAgreementDelete = "core:vendor-data-privacy-agreement:delete"

	// VendorRiskAssessment actions
	ActionVendorRiskAssessmentCreate = "core:vendor-risk-assessment:create"
	ActionVendorRiskAssessmentList   = "core:vendor-risk-assessment:list"

	// Framework actions
	ActionFrameworkGet                          = "core:framework:get"
	ActionFrameworkList                         = "core:framework:list"
	ActionFrameworkCreate                       = "core:framework:create"
	ActionFrameworkUpdate                       = "core:framework:update"
	ActionFrameworkDelete                       = "core:framework:delete"
	ActionFrameworkExport = "core:framework:export"
	ActionFrameworkImport                       = "core:framework:import"

	// Control actions
	ActionControlGet                     = "core:control:get"
	ActionControlList                    = "core:control:list"
	ActionControlCreate                  = "core:control:create"
	ActionControlUpdate                  = "core:control:update"
	ActionControlDelete                  = "core:control:delete"
	ActionControlMeasureMappingCreate    = "core:control:create-measure-mapping"
	ActionControlMeasureMappingDelete    = "core:control:delete-measure-mapping"
	ActionControlDocumentMappingCreate   = "core:control:create-document-mapping"
	ActionControlDocumentMappingDelete   = "core:control:delete-document-mapping"
	ActionControlAuditMappingCreate      = "core:control:create-audit-mapping"
	ActionControlAuditMappingDelete      = "core:control:delete-audit-mapping"
	ActionControlSnapshotMappingCreate   = "core:control:create-snapshot-mapping"
	ActionControlSnapshotMappingDelete   = "core:control:delete-snapshot-mapping"
	ActionControlObligationMappingCreate = "core:control:create-obligation-mapping"
	ActionControlObligationMappingDelete = "core:control:delete-obligation-mapping"

	// Measure actions
	ActionMeasureGet            = "core:measure:get"
	ActionMeasureList           = "core:measure:list"
	ActionMeasureCreate         = "core:measure:create"
	ActionMeasureUpdate         = "core:measure:update"
	ActionMeasureDelete         = "core:measure:delete"
	ActionMeasureEvidenceUpload = "core:measure:upload-evidence"
	ActionMeasureImport         = "core:measure:import"

	// Task actions
	ActionTaskGet      = "core:task:get"
	ActionTaskList     = "core:task:list"
	ActionTaskCreate   = "core:task:create"
	ActionTaskUpdate   = "core:task:update"
	ActionTaskDelete   = "core:task:delete"
	ActionTaskAssign   = "core:task:assign"
	ActionTaskUnassign = "core:task:unassign"

	// Evidence actions
	ActionEvidenceList   = "core:evidence:list"
	ActionEvidenceDelete = "core:evidence:delete"

	// Document actions
	ActionDocumentGet                      = "core:document:get"
	ActionDocumentList                     = "core:document:list"
	ActionDocumentCreate                   = "core:document:create"
	ActionDocumentUpdate                   = "core:document:update"
	ActionDocumentDelete                   = "core:document:delete"
	ActionDocumentChangelogGenerate        = "core:document:generate-changelog"
	ActionDocumentDraftVersionCreate       = "core:document:create-draft-version"
	ActionDocumentSendSigningNotifications = "core:document:send-signing-notifications"

	// DocumentVersion actions
	ActionDocumentVersionGet            = "core:document-version:get"
	ActionDocumentVersionList           = "core:document-version:list"
	ActionDocumentVersionExportPDF      = "core:document-version:export-pdf"
	ActionDocumentVersionExportSignable = "core:document-version:export-signable-pdf"
	ActionDocumentVersionSign           = "core:document-version:sign"
	ActionDocumentVersionUpdate         = "core:document-version:update"
	ActionDocumentVersionDeleteDraft    = "core:document-version:delete-draft"
	ActionDocumentVersionPublish        = "core:document-version:publish"
	ActionDocumentVersionExport         = "core:document-version:export"

	// DocumentVersionSignature actions
	ActionDocumentVersionSignatureRequest = "core:document-version-signature:request"
	ActionDocumentVersionCancelSignature  = "core:document-version-signature:cancel"
	ActionDocumentVersionSignatureGet     = "core:document-version-signature:get"
	ActionDocumentVersionSignatureList    = "core:document-version-signature:list"

	// Risk actions
	ActionRiskGet                     = "core:risk:get"
	ActionRiskList                    = "core:risk:list"
	ActionRiskCreate                  = "core:risk:create"
	ActionRiskUpdate                  = "core:risk:update"
	ActionRiskDelete                  = "core:risk:delete"
	ActionRiskMeasureMappingCreate    = "core:risk:create-measure-mapping"
	ActionRiskMeasureMappingDelete    = "core:risk:delete-measure-mapping"
	ActionRiskDocumentMappingCreate   = "core:risk:create-document-mapping"
	ActionRiskDocumentMappingDelete   = "core:risk:delete-document-mapping"
	ActionRiskObligationMappingCreate = "core:risk:create-obligation-mapping"
	ActionRiskObligationMappingDelete = "core:risk:delete-obligation-mapping"

	// Asset actions
	ActionAssetGet    = "core:asset:get"
	ActionAssetList   = "core:asset:list"
	ActionAssetCreate = "core:asset:create"
	ActionAssetUpdate = "core:asset:update"
	ActionAssetDelete = "core:asset:delete"

	// Datum actions
	ActionDatumGet    = "core:datum:get"
	ActionDatumList   = "core:datum:list"
	ActionDatumCreate = "core:datum:create"
	ActionDatumUpdate = "core:datum:update"
	ActionDatumDelete = "core:datum:delete"

	// Audit actions
	ActionAuditGet          = "core:audit:get"
	ActionAuditList         = "core:audit:list"
	ActionAuditCreate       = "core:audit:create"
	ActionAuditUpdate       = "core:audit:update"
	ActionAuditDelete       = "core:audit:delete"
	ActionAuditReportUpload = "core:audit:upload-report"
	ActionAuditReportDelete = "core:audit:delete-report"

	// Report actions
	ActionReportGet            = "core:report:get"
	ActionReportGetReportUrl   = "core:report:get-report-url"
	ActionReportDownloadUrlGet = "core:report:get-download-url"

	// Nonconformity actions
	ActionNonconformityGet    = "core:nonconformity:get"
	ActionNonconformityList   = "core:nonconformity:list"
	ActionNonconformityCreate = "core:nonconformity:create"
	ActionNonconformityUpdate = "core:nonconformity:update"
	ActionNonconformityDelete = "core:nonconformity:delete"

	// Obligation actions
	ActionObligationGet    = "core:obligation:get"
	ActionObligationList   = "core:obligation:list"
	ActionObligationCreate = "core:obligation:create"
	ActionObligationUpdate = "core:obligation:update"
	ActionObligationDelete = "core:obligation:delete"

	// ContinualImprovement actions
	ActionContinualImprovementGet    = "core:continual-improvement:get"
	ActionContinualImprovementList   = "core:continual-improvement:list"
	ActionContinualImprovementCreate = "core:continual-improvement:create"
	ActionContinualImprovementUpdate = "core:continual-improvement:update"
	ActionContinualImprovementDelete = "core:continual-improvement:delete"

	// ProcessingActivity actions
	ActionProcessingActivityList   = "core:processing-activity:list"
	ActionProcessingActivityGet    = "core:processing-activity:get"
	ActionProcessingActivityCreate = "core:processing-activity:create"
	ActionProcessingActivityUpdate = "core:processing-activity:update"
	ActionProcessingActivityDelete = "core:processing-activity:delete"
	ActionProcessingActivityExport = "core:processing-activity:export"

	// Snapshot actions
	ActionSnapshotGet    = "core:snapshot:get"
	ActionSnapshotList   = "core:snapshot:list"
	ActionSnapshotCreate = "core:snapshot:create"
	ActionSnapshotDelete = "core:snapshot:delete"

	// CustomDomain actions
	ActionCustomDomainGet    = "core:custom-domain:get"
	ActionCustomDomainCreate = "core:custom-domain:create"
	ActionCustomDomainDelete = "core:custom-domain:delete"

	// File actions
	ActionFileGet         = "core:file:get"
	ActionFileDownloadUrl = "core:file:download-url"

	// Meeting actions
	ActionMeetingList   = "core:meeting:list"
	ActionMeetingGet    = "core:meeting:get"
	ActionMeetingCreate = "core:meeting:create"
	ActionMeetingUpdate = "core:meeting:update"
	ActionMeetingDelete = "core:meeting:delete"

	// Connector actions
	ActionConnectorInitiate = "core:connector:initiate"

	// SlackConnection actions
	ActionSlackConnectionList = "core:slack-connection:list"

	// Connector actions (generic)
	ActionConnectorList   = "core:connector:list"
	ActionConnectorDelete = "core:connector:delete"

	// DataProtectionImpactAssessment actions
	ActionDataProtectionImpactAssessmentList   = "core:data-protection-impact-assessment:list"
	ActionDataProtectionImpactAssessmentGet    = "core:data-protection-impact-assessment:get"
	ActionDataProtectionImpactAssessmentCreate = "core:data-protection-impact-assessment:create"
	ActionDataProtectionImpactAssessmentUpdate = "core:data-protection-impact-assessment:update"
	ActionDataProtectionImpactAssessmentDelete = "core:data-protection-impact-assessment:delete"
	ActionDataProtectionImpactAssessmentExport = "core:data-protection-impact-assessment:export"

	// TransferImpactAssessment actions
	ActionTransferImpactAssessmentList   = "core:transfer-impact-assessment:list"
	ActionTransferImpactAssessmentGet    = "core:transfer-impact-assessment:get"
	ActionTransferImpactAssessmentCreate = "core:transfer-impact-assessment:create"
	ActionTransferImpactAssessmentUpdate = "core:transfer-impact-assessment:update"
	ActionTransferImpactAssessmentDelete = "core:transfer-impact-assessment:delete"
	ActionTransferImpactAssessmentExport = "core:transfer-impact-assessment:export"

	// TrustCenterDocumentAccess actions
	ActionTrustCenterDocumentAccessList = "core:trust-center-document-access:list"

	// RightsRequest actions
	ActionRightsRequestList   = "core:rights-request:list"
	ActionRightsRequestGet    = "core:rights-request:get"
	ActionRightsRequestCreate = "core:rights-request:create"
	ActionRightsRequestUpdate = "core:rights-request:update"
	ActionRightsRequestDelete = "core:rights-request:delete"

	// StateOfApplicability actions
	ActionStateOfApplicabilityList   = "core:state-of-applicability:list"
	ActionStateOfApplicabilityGet    = "core:state-of-applicability:get"
	ActionStateOfApplicabilityCreate = "core:state-of-applicability:create"
	ActionStateOfApplicabilityUpdate = "core:state-of-applicability:update"
	ActionStateOfApplicabilityDelete = "core:state-of-applicability:delete"
	ActionStateOfApplicabilityExport = "core:state-of-applicability:export"

	ActionApplicabilityStatementGet    = "core:applicability-statement:get"
	ActionApplicabilityStatementList   = "core:applicability-statement:list"
	ActionApplicabilityStatementCreate = "core:applicability-statement:create"
	ActionApplicabilityStatementUpdate = "core:applicability-statement:update"
	ActionApplicabilityStatementDelete = "core:applicability-statement:delete"

	// WebhookSubscription actions
	ActionWebhookSubscriptionList   = "core:webhook-subscription:list"
	ActionWebhookSubscriptionGet    = "core:webhook-subscription:get"
	ActionWebhookSubscriptionCreate = "core:webhook-subscription:create"
	ActionWebhookSubscriptionUpdate = "core:webhook-subscription:update"
	ActionWebhookSubscriptionDelete = "core:webhook-subscription:delete"
)
