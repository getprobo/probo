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

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var (
	organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")
	ownerCondition        = policy.Equals("principal.id", "resource.owner_id")
)

// OwnerPolicy defines permissions for organization owners.
var OwnerPolicy = policy.NewPolicy(
	"probo:owner",
	"Probo Owner",
	policy.Allow("core:*").WithSID("full-core-access").When(organizationCondition),
).WithDescription("Full probo access for organization owners")

// AdminPolicy defines permissions for organization admins.
var AdminPolicy = policy.NewPolicy(
	"probo:admin",
	"Probo Admin",
	policy.Allow("core:*").WithSID("full-core-access").When(organizationCondition),
).WithDescription("Probo admin access - can manage core entities")

// ViewerPolicy defines read-only permissions for organization viewers.
var ViewerPolicy = policy.NewPolicy(
	"probo:viewer",
	"Probo Viewer",
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
		ActionOrganizationGetHorizontalLogoUrl,
	).WithSID("org-read-access").When(organizationCondition),

	policy.Allow(
		ActionThirdPartyGet, ActionThirdPartyList,
		ActionThirdPartyContactGet, ActionThirdPartyContactList,
		ActionThirdPartyServiceGet, ActionThirdPartyServiceList,
		ActionThirdPartyComplianceReportGet, ActionThirdPartyComplianceReportList,
		ActionThirdPartyBusinessAssociateAgreementGet,
		ActionThirdPartyDataPrivacyAgreementGet,
		ActionThirdPartyRiskAssessmentList,
		ActionThirdPartyRelationList,
		ActionFrameworkGet, ActionFrameworkList,
		ActionControlGet, ActionControlList,
		ActionMeasureGet, ActionMeasureList,
		ActionTaskGet, ActionTaskList,
		ActionTaskCommentGet, ActionTaskCommentList,
		ActionEvidenceList,
		ActionDocumentGet, ActionDocumentList,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSignatureGet, ActionDocumentVersionSignatureList,
		ActionDocumentVersionApprovalList,
		ActionElectronicSignatureGet,
		ActionRiskGet, ActionRiskList,
		ActionAssetGet, ActionAssetList,
		ActionDatumGet, ActionDatumList,
		ActionAuditGet, ActionAuditList,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionFindingGet, ActionFindingList,
		ActionObligationGet, ActionObligationList,
		ActionBusinessFunctionGet, ActionBusinessFunctionList,
		ActionAiSystemGet, ActionAiSystemList,
		ActionProcessingActivityGet, ActionProcessingActivityList,
		ActionDataProtectionImpactAssessmentGet, ActionDataProtectionImpactAssessmentList,
		ActionTransferImpactAssessmentGet, ActionTransferImpactAssessmentList,
		ActionFileGet,
		ActionSlackConnectionList, ActionConnectorList,
		ActionRightsRequestGet, ActionRightsRequestList,
		ActionStatementOfApplicabilityGet, ActionStatementOfApplicabilityList,
		ActionApplicabilityStatementGet, ActionApplicabilityStatementList,
		ActionWebhookSubscriptionGet, ActionWebhookSubscriptionList,
		ActionCookieBannerGet, ActionCookieBannerList,
		ActionCookieBannerVersionGet, ActionCookieBannerVersionList,
		ActionCookieCategoryGet, ActionCookieCategoryList,
		ActionCookieGet, ActionCookieList,
		ActionCookieConsentRecordList,
	).WithSID("entity-read-access").When(organizationCondition),

	policy.Allow(ActionOrganizationContextGet).WithSID("organization-context-read").When(organizationCondition),
	policy.Allow(
		ActionDocumentVersionExportPDF, ActionDocumentVersionSign,
	).WithSID("document-signing").When(organizationCondition),

	policy.Allow(
		ActionDocumentVersionApprove, ActionDocumentVersionReject,
	).WithSID("document-approval").When(organizationCondition),

	policy.Allow(
		ActionEmployeeDocumentGet, ActionEmployeeDocumentList,
		ActionEmployeeDocumentVersionExportPDF,
	).WithSID("employee-document-access").When(organizationCondition),
).WithDescription("Read-only probo access for organization viewers")

// AuditorPolicy defines permissions for auditor role.
var AuditorPolicy = policy.NewPolicy(
	"probo:auditor",
	"Probo Auditor",
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
		ActionOrganizationGetHorizontalLogoUrl,
	).WithSID("org-read-access").When(organizationCondition),

	policy.Allow(ActionOrganizationContextGet).WithSID("organization-context-read").When(organizationCondition),

	policy.Allow(
		ActionThirdPartyGet, ActionThirdPartyList,
		ActionThirdPartyContactGet, ActionThirdPartyContactList,
		ActionThirdPartyServiceGet, ActionThirdPartyServiceList,
		ActionThirdPartyComplianceReportGet, ActionThirdPartyComplianceReportList,
		ActionThirdPartyBusinessAssociateAgreementGet,
		ActionThirdPartyDataPrivacyAgreementGet,
		ActionThirdPartyRiskAssessmentList,
		ActionThirdPartyRelationList,
		ActionFrameworkGet, ActionFrameworkList,
		ActionControlGet, ActionControlList,
		ActionMeasureGet, ActionMeasureList,
		ActionEvidenceList,
		ActionDocumentGet, ActionDocumentList,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSignatureGet, ActionDocumentVersionSignatureList,
		ActionDocumentVersionApprovalList,
		ActionElectronicSignatureGet,
		ActionRiskGet, ActionRiskList,
		ActionAssetGet, ActionAssetList,
		ActionDatumGet, ActionDatumList,
		ActionAuditGet, ActionAuditList,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionFindingGet, ActionFindingList,
		ActionObligationGet, ActionObligationList,
		ActionBusinessFunctionGet, ActionBusinessFunctionList,
		ActionAiSystemGet, ActionAiSystemList,
		ActionProcessingActivityGet, ActionProcessingActivityList,
		ActionDataProtectionImpactAssessmentGet, ActionDataProtectionImpactAssessmentList,
		ActionTransferImpactAssessmentGet, ActionTransferImpactAssessmentList,
		ActionFileGet,
		ActionStatementOfApplicabilityGet, ActionStatementOfApplicabilityList,
		ActionApplicabilityStatementGet, ActionApplicabilityStatementList,
	).WithSID("entity-read-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentVersionExportPDF, ActionDocumentVersionSign,
	).WithSID("document-signing").When(organizationCondition),

	policy.Allow(
		ActionEmployeeDocumentGet, ActionEmployeeDocumentList,
		ActionEmployeeDocumentVersionExportPDF,
	).WithSID("employee-document-access").When(organizationCondition),
).WithDescription("Read-only probo access for auditors (excludes internal/employee content)")

// TaskCommentOwnershipPolicy is attached to owner, admin, and viewer so
// permission(action:) on a comment matches mutation authorization: only the
// author can update; the author can delete; owner/admin still delete any
// comment through core:* on their role policy.
var TaskCommentOwnershipPolicy = policy.NewPolicy(
	"probo:task-comment-ownership",
	"Task Comment Ownership",
	policy.Deny(ActionTaskCommentUpdate).
		WithSID("deny-update-others-task-comments").
		When(policy.NotEquals("principal.id", "resource.owner_id")),
	policy.Allow(ActionTaskCommentUpdate, ActionTaskCommentDelete).
		WithSID("manage-own-task-comment").
		When(organizationCondition, ownerCondition),
).WithDescription("Authors can update and delete their own task comments; nobody else can update them")

// CommonThirdPartyCatalogPolicy grants every authenticated identity
// read access to the global common third-party catalog. The catalog is
// shared across all tenants and has no organization scoping, so the
// allow has no condition.
var CommonThirdPartyCatalogPolicy = policy.NewPolicy(
	"probo:common-third-party-catalog",
	"Probo Common Third-Party Catalog",
	policy.Allow(
		ActionCommonThirdPartyGet,
		ActionCommonThirdPartyList,
	).WithSID("read-common-third-party-catalog"),
).WithDescription("Allows every authenticated user to read the global common third-party catalog")

// EmployeePolicy defines permissions for employee role.
var EmployeePolicy = policy.NewPolicy(
	"probo:employee",
	"Probo Employee",
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
	).WithSID("org-basic-access").When(organizationCondition),

	policy.Allow(
		ActionEmployeeDocumentGet, ActionEmployeeDocumentList,
	).WithSID("employee-document-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentVersionSign,
		ActionEmployeeDocumentVersionExportPDF,
	).WithSID("document-version-signing").When(organizationCondition),

	policy.Allow(
		ActionDocumentVersionApprovalList,
		ActionDocumentVersionApprove,
		ActionDocumentVersionReject,
	).WithSID("document-version-approval").When(organizationCondition),
).WithDescription("Employee access - can sign documents, approve documents, and view internal content")

// CompliancePortalManagerPolicy defines permissions needed to manage the
// compliance portal and toggle portal visibility on related core entities.
var CompliancePortalManagerPolicy = policy.NewPolicy(
	"probo:compliance-portal-manager",
	"Probo Compliance Portal Manager",
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
		ActionOrganizationGetHorizontalLogoUrl,
	).WithSID("org-read-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentGet, ActionDocumentList, ActionDocumentUpdate,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionAuditGet, ActionAuditList, ActionAuditUpdate,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionFrameworkGet, ActionFrameworkList,
		ActionThirdPartyGet, ActionThirdPartyList, ActionThirdPartyUpdate,
		ActionFileGet,
		ActionElectronicSignatureGet,
		ActionSlackConnectionList, ActionConnectorList,
		ActionConnectorInitiate, ActionConnectorDelete,
	).WithSID("compliance-portal-related-access").When(organizationCondition),
).WithDescription("Access required to manage the compliance portal and related entity visibility")

// CompliancePortalAccessManagerPolicy defines permissions needed to review and
// approve compliance portal visitor access requests.
//
// Related document, audit, report, and file reads are intentionally
// organization-scoped: access requests reference those entities by ID, and the
// authorizer has no request-entitlement condition yet. Callers still only load
// the IDs attached to requests they can list.
var CompliancePortalAccessManagerPolicy = policy.NewPolicy(
	"probo:compliance-portal-access-manager",
	"Probo Compliance Portal Access Manager",
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
		ActionOrganizationGetHorizontalLogoUrl,
	).WithSID("org-read-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentGet,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionAuditGet,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionFrameworkGet,
		ActionFileGet,
		ActionElectronicSignatureGet,
	).WithSID("compliance-portal-access-related").When(organizationCondition),
).WithDescription("Organization-scoped read of entities that portal access requests may reference")

// ProboPolicySet returns the PolicySet for the probo service.
func ProboPolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", OwnerPolicy, TaskCommentOwnershipPolicy).
		AddRolePolicy("ADMIN", AdminPolicy, TaskCommentOwnershipPolicy).
		AddRolePolicy("VIEWER", ViewerPolicy, TaskCommentOwnershipPolicy).
		AddRolePolicy("AUDITOR", AuditorPolicy).
		AddRolePolicy("EMPLOYEE", EmployeePolicy).
		AddRolePolicy("COMPLIANCE_PORTAL_MANAGER", CompliancePortalManagerPolicy).
		AddRolePolicy("COMPLIANCE_PORTAL_ACCESS_MANAGER", CompliancePortalAccessManagerPolicy).
		AddIdentityScopedPolicy(CommonThirdPartyCatalogPolicy)
}
