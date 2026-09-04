// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

	// ThirdParty actions
	ActionThirdPartyList    = "core:thirdParty:list"
	ActionThirdPartyGet     = "core:thirdParty:get"
	ActionThirdPartyCreate  = "core:thirdParty:create"
	ActionThirdPartyUpdate  = "core:thirdParty:update"
	ActionThirdPartyDelete  = "core:thirdParty:delete"
	ActionThirdPartyVet     = "core:thirdParty:vet"
	ActionThirdPartyPublish = "core:thirdParty:publish"

	// ThirdPartyRelation actions
	ActionThirdPartyRelationCreate = "core:thirdParty-relation:create"
	ActionThirdPartyRelationList   = "core:thirdParty-relation:list"

	// ThirdPartyContact actions
	ActionThirdPartyContactGet    = "core:thirdParty-contact:get"
	ActionThirdPartyContactList   = "core:thirdParty-contact:list"
	ActionThirdPartyContactCreate = "core:thirdParty-contact:create"
	ActionThirdPartyContactUpdate = "core:thirdParty-contact:update"
	ActionThirdPartyContactDelete = "core:thirdParty-contact:delete"

	// ThirdPartyService actions
	ActionThirdPartyServiceGet    = "core:thirdParty-service:get"
	ActionThirdPartyServiceList   = "core:thirdParty-service:list"
	ActionThirdPartyServiceCreate = "core:thirdParty-service:create"
	ActionThirdPartyServiceUpdate = "core:thirdParty-service:update"
	ActionThirdPartyServiceDelete = "core:thirdParty-service:delete"

	// ThirdPartyComplianceReport actions
	ActionThirdPartyComplianceReportGet    = "core:thirdParty-compliance-report:get"
	ActionThirdPartyComplianceReportList   = "core:thirdParty-compliance-report:list"
	ActionThirdPartyComplianceReportUpload = "core:thirdParty-compliance-report:upload"
	ActionThirdPartyComplianceReportDelete = "core:thirdParty-compliance-report:delete"

	// ThirdPartyBusinessAssociateAgreement actions
	ActionThirdPartyBusinessAssociateAgreementGet    = "core:thirdParty-business-associate-agreement:get"
	ActionThirdPartyBusinessAssociateAgreementUpload = "core:thirdParty-business-associate-agreement:upload"
	ActionThirdPartyBusinessAssociateAgreementUpdate = "core:thirdParty-business-associate-agreement:update"
	ActionThirdPartyBusinessAssociateAgreementDelete = "core:thirdParty-business-associate-agreement:delete"

	// ThirdPartyDataPrivacyAgreement actions
	ActionThirdPartyDataPrivacyAgreementGet    = "core:thirdParty-data-privacy-agreement:get"
	ActionThirdPartyDataPrivacyAgreementUpload = "core:thirdParty-data-privacy-agreement:upload"
	ActionThirdPartyDataPrivacyAgreementUpdate = "core:thirdParty-data-privacy-agreement:update"
	ActionThirdPartyDataPrivacyAgreementDelete = "core:thirdParty-data-privacy-agreement:delete"

	// ThirdPartyRiskAssessment actions
	ActionThirdPartyRiskAssessmentCreate = "core:thirdParty-risk-assessment:create"
	ActionThirdPartyRiskAssessmentList   = "core:thirdParty-risk-assessment:list"

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
	ActionControlObligationMappingCreate = "core:control:create-obligation-mapping"
	ActionControlObligationMappingDelete = "core:control:delete-obligation-mapping"

	// Measure actions
	ActionMeasureGet                     = "core:measure:get"
	ActionMeasureList                    = "core:measure:list"
	ActionMeasureCreate                  = "core:measure:create"
	ActionMeasureUpdate                  = "core:measure:update"
	ActionMeasureDelete                  = "core:measure:delete"
	ActionMeasureEvidenceUpload          = "core:measure:upload-evidence"
	ActionMeasureImport                  = "core:measure:import"
	ActionMeasureDocumentMappingCreate   = "core:measure:create-document-mapping"
	ActionMeasureDocumentMappingDelete   = "core:measure:delete-document-mapping"
	ActionMeasureThirdPartyMappingCreate = "core:measure:create-third-party-mapping"
	ActionMeasureThirdPartyMappingDelete = "core:measure:delete-third-party-mapping"

	// Task actions
	ActionTaskGet      = "core:task:get"
	ActionTaskList     = "core:task:list"
	ActionTaskCreate   = "core:task:create"
	ActionTaskUpdate   = "core:task:update"
	ActionTaskDelete   = "core:task:delete"
	ActionTaskAssign   = "core:task:assign"
	ActionTaskUnassign = "core:task:unassign"

	// TaskComment actions
	ActionTaskCommentGet    = "core:task-comment:get"
	ActionTaskCommentList   = "core:task-comment:list"
	ActionTaskCommentCreate = "core:task-comment:create"
	ActionTaskCommentUpdate = "core:task-comment:update"
	ActionTaskCommentDelete = "core:task-comment:delete"

	// Evidence actions
	ActionEvidenceList   = "core:evidence:list"
	ActionEvidenceDelete = "core:evidence:delete"

	// Document actions
	ActionDocumentGet         = "core:document:get"
	ActionDocumentList        = "core:document:list"
	ActionDocumentCreate      = "core:document:create"
	ActionDocumentUpdate      = "core:document:update"
	ActionDocumentDelete      = "core:document:delete"
	ActionDocumentArchive     = "core:document:archive"
	ActionDocumentUnarchive   = "core:document:unarchive"
	ActionDocumentDeleteDraft = "core:document:delete-draft"

	// DocumentVersion actions
	ActionDocumentVersionGet             = "core:document-version:get"
	ActionDocumentVersionList            = "core:document-version:list"
	ActionDocumentVersionExportPDF       = "core:document-version:export-pdf"
	ActionDocumentVersionSign            = "core:document-version:sign"
	ActionDocumentVersionRequestApproval = "core:document-version:request-approval"
	ActionDocumentVersionVoidApproval    = "core:document-version:void-approval"
	ActionDocumentVersionApprove         = "core:document-version:approve"
	ActionDocumentVersionReject          = "core:document-version:reject"
	ActionDocumentVersionApprovalList    = "core:document-version:approval-list"
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
	ActionRiskGet                     = "risk-management:risk:get"
	ActionRiskList                    = "risk-management:risk:list"
	ActionRiskCreate                  = "risk-management:risk:create"
	ActionRiskUpdate                  = "risk-management:risk:update"
	ActionRiskDelete                  = "risk-management:risk:delete"
	ActionRiskMeasureMappingCreate    = "risk-management:risk:create-measure-mapping"
	ActionRiskMeasureMappingDelete    = "risk-management:risk:delete-measure-mapping"
	ActionRiskDocumentMappingCreate   = "risk-management:risk:create-document-mapping"
	ActionRiskDocumentMappingDelete   = "risk-management:risk:delete-document-mapping"
	ActionRiskObligationMappingCreate = "risk-management:risk:create-obligation-mapping"
	ActionRiskObligationMappingDelete = "risk-management:risk:delete-obligation-mapping"
	ActionRiskPublish                 = "risk-management:risk:publish"

	// Asset actions
	ActionAssetGet     = "core:asset:get"
	ActionAssetList    = "core:asset:list"
	ActionAssetCreate  = "core:asset:create"
	ActionAssetUpdate  = "core:asset:update"
	ActionAssetDelete  = "core:asset:delete"
	ActionAssetPublish = "core:asset:publish"

	// Datum actions
	ActionDatumGet     = "core:datum:get"
	ActionDatumList    = "core:datum:list"
	ActionDatumCreate  = "core:datum:create"
	ActionDatumUpdate  = "core:datum:update"
	ActionDatumDelete  = "core:datum:delete"
	ActionDatumPublish = "core:datum:publish"

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
	ActionFindingPublish            = "core:finding:publish"

	// Obligation actions
	ActionObligationGet     = "core:obligation:get"
	ActionObligationList    = "core:obligation:list"
	ActionObligationCreate  = "core:obligation:create"
	ActionObligationUpdate  = "core:obligation:update"
	ActionObligationDelete  = "core:obligation:delete"
	ActionObligationPublish = "core:obligation:publish"

	// BusinessFunction actions
	ActionBusinessFunctionGet     = "core:business-function:get"
	ActionBusinessFunctionList    = "core:business-function:list"
	ActionBusinessFunctionCreate  = "core:business-function:create"
	ActionBusinessFunctionUpdate  = "core:business-function:update"
	ActionBusinessFunctionDelete  = "core:business-function:delete"
	ActionBusinessFunctionPublish = "core:business-function:publish"

	// AiSystem actions
	ActionAiSystemGet     = "core:ai-system:get"
	ActionAiSystemList    = "core:ai-system:list"
	ActionAiSystemCreate  = "core:ai-system:create"
	ActionAiSystemUpdate  = "core:ai-system:update"
	ActionAiSystemDelete  = "core:ai-system:delete"
	ActionAiSystemPublish = "core:ai-system:publish"

	// ProcessingActivity actions
	ActionProcessingActivityList    = "core:processing-activity:list"
	ActionProcessingActivityGet     = "core:processing-activity:get"
	ActionProcessingActivityCreate  = "core:processing-activity:create"
	ActionProcessingActivityUpdate  = "core:processing-activity:update"
	ActionProcessingActivityDelete  = "core:processing-activity:delete"
	ActionProcessingActivityPublish = "core:processing-activity:publish"

	// File actions
	ActionFileGet = "core:file:get"

	// Connector actions
	ActionConnectorInitiate = "core:connector:initiate"

	// SlackConnection actions
	ActionSlackConnectionList = "core:slack-connection:list"

	// Connector actions (generic)
	ActionConnectorCreate = "core:connector:create"
	ActionConnectorGet    = "core:connector:get"
	ActionConnectorList   = "core:connector:list"
	ActionConnectorDelete = "core:connector:delete"

	// DataProtectionImpactAssessment actions
	ActionDataProtectionImpactAssessmentList    = "core:data-protection-impact-assessment:list"
	ActionDataProtectionImpactAssessmentGet     = "core:data-protection-impact-assessment:get"
	ActionDataProtectionImpactAssessmentCreate  = "core:data-protection-impact-assessment:create"
	ActionDataProtectionImpactAssessmentUpdate  = "core:data-protection-impact-assessment:update"
	ActionDataProtectionImpactAssessmentDelete  = "core:data-protection-impact-assessment:delete"
	ActionDataProtectionImpactAssessmentPublish = "core:data-protection-impact-assessment:publish"

	// TransferImpactAssessment actions
	ActionTransferImpactAssessmentList    = "core:transfer-impact-assessment:list"
	ActionTransferImpactAssessmentGet     = "core:transfer-impact-assessment:get"
	ActionTransferImpactAssessmentCreate  = "core:transfer-impact-assessment:create"
	ActionTransferImpactAssessmentUpdate  = "core:transfer-impact-assessment:update"
	ActionTransferImpactAssessmentDelete  = "core:transfer-impact-assessment:delete"
	ActionTransferImpactAssessmentPublish = "core:transfer-impact-assessment:publish"

	// RightsRequest actions
	ActionRightsRequestList   = "core:rights-request:list"
	ActionRightsRequestGet    = "core:rights-request:get"
	ActionRightsRequestCreate = "core:rights-request:create"
	ActionRightsRequestUpdate = "core:rights-request:update"
	ActionRightsRequestDelete = "core:rights-request:delete"

	// StatementOfApplicability actions
	ActionStatementOfApplicabilityList    = "core:statement-of-applicability:list"
	ActionStatementOfApplicabilityGet     = "core:statement-of-applicability:get"
	ActionStatementOfApplicabilityCreate  = "core:statement-of-applicability:create"
	ActionStatementOfApplicabilityUpdate  = "core:statement-of-applicability:update"
	ActionStatementOfApplicabilityDelete  = "core:statement-of-applicability:delete"
	ActionStatementOfApplicabilityPublish = "core:statement-of-applicability:publish"

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

	// CookieBanner actions
	ActionCookieBannerGet        = "core:cookie-banner:get"
	ActionCookieBannerList       = "core:cookie-banner:list"
	ActionCookieBannerCreate     = "core:cookie-banner:create"
	ActionCookieBannerUpdate     = "core:cookie-banner:update"
	ActionCookieBannerDelete     = "core:cookie-banner:delete"
	ActionCookieBannerActivate   = "core:cookie-banner:activate"
	ActionCookieBannerDeactivate = "core:cookie-banner:deactivate"

	ActionCookieBannerRegeneratePolicy = "core:cookie-banner:regenerate-policy"

	// CookieBannerVersion actions
	ActionCookieBannerVersionGet     = "core:cookie-banner-version:get"
	ActionCookieBannerVersionList    = "core:cookie-banner-version:list"
	ActionCookieBannerVersionPublish = "core:cookie-banner-version:publish"

	// CookieCategory actions
	ActionCookieCategoryGet    = "core:cookie-category:get"
	ActionCookieCategoryList   = "core:cookie-category:list"
	ActionCookieCategoryCreate = "core:cookie-category:create"
	ActionCookieCategoryUpdate = "core:cookie-category:update"
	ActionCookieCategoryDelete = "core:cookie-category:delete"

	// Cookie actions
	ActionCookieGet    = "core:cookie:get"
	ActionCookieList   = "core:cookie:list"
	ActionCookieCreate = "core:cookie:create"
	ActionCookieUpdate = "core:cookie:update"
	ActionCookieDelete = "core:cookie:delete"

	// TrackerPattern actions
	ActionTrackerPatternGet    = "core:tracker-pattern:get"
	ActionTrackerPatternList   = "core:tracker-pattern:list"
	ActionTrackerPatternCreate = "core:tracker-pattern:create"
	ActionTrackerPatternUpdate = "core:tracker-pattern:update"
	ActionTrackerPatternDelete = "core:tracker-pattern:delete"

	// TrackerResource actions
	ActionTrackerResourceGet    = "core:tracker-resource:get"
	ActionTrackerResourceList   = "core:tracker-resource:list"
	ActionTrackerResourceCreate = "core:tracker-resource:create"
	ActionTrackerResourceUpdate = "core:tracker-resource:update"
	ActionTrackerResourceDelete = "core:tracker-resource:delete"

	// CookieConsentRecord actions
	ActionCookieConsentRecordList = "core:cookie-consent-record:list"

	// CommonThirdParty actions (global catalog, no organization scope).
	ActionCommonThirdPartyGet  = "core:common-third-party:get"
	ActionCommonThirdPartyList = "core:common-third-party:list"

	// CommonGVLVendor actions (global IAB GVL catalog, no organization scope).
	ActionCommonGVLVendorList = "core:common-gvl-vendor:list"

	// ElectronicSignature actions (tenant-scoped via the related document
	// version signature / compliance portal access).
	ActionElectronicSignatureGet = "core:electronic-signature:get"
)
