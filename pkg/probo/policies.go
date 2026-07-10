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
		ActionRiskAssessmentGet, ActionRiskAssessmentList,
		ActionRiskAssessmentScopeGet, ActionRiskAssessmentScopeList,
		ActionRiskAssessmentNodeGet, ActionRiskAssessmentNodeList,
		ActionRiskAssessmentBoundaryGet, ActionRiskAssessmentBoundaryList,
		ActionRiskAssessmentProcessGet, ActionRiskAssessmentProcessList,
		ActionRiskAssessmentThreatGet, ActionRiskAssessmentThreatList,
		ActionRiskAssessmentScenarioGet, ActionRiskAssessmentScenarioList,
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
		ActionProcessingActivityGet, ActionProcessingActivityList,
		ActionDataProtectionImpactAssessmentGet, ActionDataProtectionImpactAssessmentList,
		ActionTransferImpactAssessmentGet, ActionTransferImpactAssessmentList,
		ActionFileGet,
		ActionStatementOfApplicabilityGet, ActionStatementOfApplicabilityList,
		ActionApplicabilityStatementGet, ActionApplicabilityStatementList,
		ActionRiskAssessmentGet, ActionRiskAssessmentList,
		ActionRiskAssessmentScopeGet, ActionRiskAssessmentScopeList,
		ActionRiskAssessmentNodeGet, ActionRiskAssessmentNodeList,
		ActionRiskAssessmentBoundaryGet, ActionRiskAssessmentBoundaryList,
		ActionRiskAssessmentProcessGet, ActionRiskAssessmentProcessList,
		ActionRiskAssessmentThreatGet, ActionRiskAssessmentThreatList,
		ActionRiskAssessmentScenarioGet, ActionRiskAssessmentScenarioList,
	).WithSID("entity-read-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentVersionExportPDF, ActionDocumentVersionSign,
	).WithSID("document-signing").When(organizationCondition),

	policy.Allow(
		ActionEmployeeDocumentGet, ActionEmployeeDocumentList,
		ActionEmployeeDocumentVersionExportPDF,
	).WithSID("employee-document-access").When(organizationCondition),
).WithDescription("Read-only probo access for auditors (excludes internal/employee content)")

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

// ProboPolicySet returns the PolicySet for the probo service.
func ProboPolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", OwnerPolicy).
		AddRolePolicy("ADMIN", AdminPolicy).
		AddRolePolicy("VIEWER", ViewerPolicy).
		AddRolePolicy("AUDITOR", AuditorPolicy).
		AddRolePolicy("EMPLOYEE", EmployeePolicy).
		AddIdentityScopedPolicy(CommonThirdPartyCatalogPolicy)
}
