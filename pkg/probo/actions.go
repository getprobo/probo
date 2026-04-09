// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	ActionOrganizationUpdate               = "core:organization:update"
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

	// MailingListUpdate actions
	ActionMailingListUpdateList   = "core:mailing-list-update:list"
	ActionMailingListUpdateCreate = "core:mailing-list-update:create"
	ActionMailingListUpdateUpdate = "core:mailing-list-update:update"
	ActionMailingListUpdateSend   = "core:mailing-list-update:send"
	ActionMailingListUpdateDelete = "core:mailing-list-update:delete"

	// MailingList actions
	ActionMailingListUpdate = "core:mailing-list:update"

	// MailingListSubscriber actions
	ActionMailingListSubscriberList   = "core:mailing-list-subscriber:list"
	ActionMailingListSubscriberCreate = "core:mailing-list-subscriber:create"
	ActionMailingListSubscriberDelete = "core:mailing-list-subscriber:delete"

	// TrustCenterReference actions
	ActionTrustCenterReferenceList       = "core:trust-center-reference:list"
	ActionTrustCenterReferenceGetLogoUrl = "core:trust-center-reference:get-logo-url"
	ActionTrustCenterReferenceCreate     = "core:trust-center-reference:create"
	ActionTrustCenterReferenceUpdate     = "core:trust-center-reference:update"
	ActionTrustCenterReferenceDelete     = "core:trust-center-reference:delete"

	// ComplianceFramework actions
	ActionComplianceFrameworkList       = "core:compliance-framework:list"
	ActionComplianceFrameworkCreate     = "core:compliance-framework:create"
	ActionComplianceFrameworkDelete     = "core:compliance-framework:delete"
	ActionComplianceFrameworkUpdateRank = "core:compliance-framework:update-rank"

	// ComplianceExternalURL actions
	ActionComplianceExternalURLList   = "core:compliance-external-url:list"
	ActionComplianceExternalURLCreate = "core:compliance-external-url:create"
	ActionComplianceExternalURLUpdate = "core:compliance-external-url:update"
	ActionComplianceExternalURLDelete = "core:compliance-external-url:delete"

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
	ActionFrameworkGet    = "core:framework:get"
	ActionFrameworkList   = "core:framework:list"
	ActionFrameworkCreate = "core:framework:create"
	ActionFrameworkUpdate = "core:framework:update"
	ActionFrameworkDelete = "core:framework:delete"
	ActionFrameworkExport = "core:framework:export"
	ActionFrameworkImport = "core:framework:import"

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
	ActionMeasureGet                   = "core:measure:get"
	ActionMeasureList                  = "core:measure:list"
	ActionMeasureCreate                = "core:measure:create"
	ActionMeasureUpdate                = "core:measure:update"
	ActionMeasureDelete                = "core:measure:delete"
	ActionMeasureEvidenceUpload        = "core:measure:upload-evidence"
	ActionMeasureImport                = "core:measure:import"
	ActionMeasureDocumentMappingCreate = "core:measure:create-document-mapping"
	ActionMeasureDocumentMappingDelete = "core:measure:delete-document-mapping"

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
	ActionDocumentArchive                  = "core:document:archive"
	ActionDocumentUnarchive                = "core:document:unarchive"
	ActionDocumentDraftVersionCreate       = "core:document:create-draft-version"
	ActionDocumentSendSigningNotifications = "core:document:send-signing-notifications"

	// DocumentVersion actions
	ActionDocumentVersionGet             = "core:document-version:get"
	ActionDocumentVersionList            = "core:document-version:list"
	ActionDocumentVersionExportPDF       = "core:document-version:export-pdf"
	ActionDocumentVersionSign            = "core:document-version:sign"
	ActionDocumentVersionUpdate          = "core:document-version:update"
	ActionDocumentVersionDeleteDraft     = "core:document-version:delete-draft"
	ActionDocumentVersionRequestApproval = "core:document-version:request-approval"
	ActionDocumentVersionApprove         = "core:document-version:approve"
	ActionDocumentVersionReject          = "core:document-version:reject"
	ActionDocumentVersionApprovalList    = "core:document-version:approval-list"
	ActionDocumentVersionAddApprover     = "core:document-version:add-approver"
	ActionDocumentVersionRemoveApprover  = "core:document-version:remove-approver"
	ActionDocumentVersionPublish         = "core:document-version:publish"
	ActionDocumentVersionExport          = "core:document-version:export"

	// EmployeeDocument actions
	ActionEmployeeDocumentGet              = "core:employee-document:get"
	ActionEmployeeDocumentList             = "core:employee-document:list"
	ActionEmployeeDocumentVersionExportPDF = "core:employee-document-version:export-pdf"

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

	// Finding actions
	ActionFindingGet                = "core:finding:get"
	ActionFindingList               = "core:finding:list"
	ActionFindingCreate             = "core:finding:create"
	ActionFindingUpdate             = "core:finding:update"
	ActionFindingDelete             = "core:finding:delete"
	ActionFindingAuditMappingCreate = "core:finding:create-audit-mapping"
	ActionFindingAuditMappingDelete = "core:finding:delete-audit-mapping"

	// Obligation actions
	ActionObligationGet    = "core:obligation:get"
	ActionObligationList   = "core:obligation:list"
	ActionObligationCreate = "core:obligation:create"
	ActionObligationUpdate = "core:obligation:update"
	ActionObligationDelete = "core:obligation:delete"

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
	ActionConnectorCreate = "core:connector:create"
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

	// AccessReviewCampaign actions
	ActionAccessReviewCampaignGet               = "core:access-review-campaign:get"
	ActionAccessReviewCampaignList              = "core:access-review-campaign:list"
	ActionAccessReviewCampaignCreate            = "core:access-review-campaign:create"
	ActionAccessReviewCampaignUpdate            = "core:access-review-campaign:update"
	ActionAccessReviewCampaignDelete            = "core:access-review-campaign:delete"
	ActionAccessReviewCampaignStart             = "core:access-review-campaign:start"
	ActionAccessReviewCampaignClose             = "core:access-review-campaign:close"
	ActionAccessReviewCampaignCancel            = "core:access-review-campaign:cancel"
	ActionAccessReviewCampaignAddScopeSource    = "core:access-review-campaign:add-scope-source"
	ActionAccessReviewCampaignRemoveScopeSource = "core:access-review-campaign:remove-scope-source"

	// AccessEntry actions
	ActionAccessEntryGet    = "core:access-entry:get"
	ActionAccessEntryList   = "core:access-entry:list"
	ActionAccessEntryDecide = "core:access-entry:decide"
	ActionAccessEntryFlag   = "core:access-entry:flag"

	// AccessSource actions
	ActionAccessSourceGet    = "core:access-source:get"
	ActionAccessSourceList   = "core:access-source:list"
	ActionAccessSourceCreate = "core:access-source:create"
	ActionAccessSourceUpdate = "core:access-source:update"
	ActionAccessSourceDelete = "core:access-source:delete"
	ActionAccessSourceSync   = "core:access-source:sync"

	// CookieBanner actions
	ActionCookieBannerGet    = "core:cookie-banner:get"
	ActionCookieBannerList   = "core:cookie-banner:list"
	ActionCookieBannerCreate = "core:cookie-banner:create"
	ActionCookieBannerUpdate = "core:cookie-banner:update"
	ActionCookieBannerDelete = "core:cookie-banner:delete"

	// CookieCategory actions
	ActionCookieCategoryGet    = "core:cookie-category:get"
	ActionCookieCategoryList   = "core:cookie-category:list"
	ActionCookieCategoryCreate = "core:cookie-category:create"
	ActionCookieCategoryUpdate = "core:cookie-category:update"
	ActionCookieCategoryDelete = "core:cookie-category:delete"

	// ConsentRecord actions
	ActionConsentRecordList = "core:consent-record:list"
)
