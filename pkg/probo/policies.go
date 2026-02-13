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

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

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
		ActionVendorGet, ActionVendorList,
		ActionVendorContactGet, ActionVendorContactList,
		ActionVendorServiceGet, ActionVendorServiceList,
		ActionVendorComplianceReportGet, ActionVendorComplianceReportList,
		ActionVendorBusinessAssociateAgreementGet,
		ActionVendorDataPrivacyAgreementGet,
		ActionVendorRiskAssessmentList,
		ActionFrameworkGet, ActionFrameworkList,
		ActionControlGet, ActionControlList,
		ActionMeasureGet, ActionMeasureList,
		ActionTaskGet, ActionTaskList,
		ActionEvidenceList,
		ActionDocumentGet, ActionDocumentList,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSignatureGet, ActionDocumentVersionSignatureList,
		ActionRiskGet, ActionRiskList,
		ActionAssetGet, ActionAssetList,
		ActionDatumGet, ActionDatumList,
		ActionAuditGet, ActionAuditList,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionNonconformityGet, ActionNonconformityList,
		ActionObligationGet, ActionObligationList,
		ActionContinualImprovementGet, ActionContinualImprovementList,
		ActionProcessingActivityGet, ActionProcessingActivityList,
		ActionDataProtectionImpactAssessmentGet, ActionDataProtectionImpactAssessmentList,
		ActionTransferImpactAssessmentGet, ActionTransferImpactAssessmentList,
		ActionSnapshotGet, ActionSnapshotList,
		ActionMeetingGet, ActionMeetingList,
		ActionFileGet, ActionFileDownloadUrl,
		ActionSlackConnectionList, ActionConnectorList,
		ActionRightsRequestGet, ActionRightsRequestList,
		ActionStateOfApplicabilityGet, ActionStateOfApplicabilityList,
		ActionApplicabilityStatementGet, ActionApplicabilityStatementList,
		ActionWebhookSubscriptionGet, ActionWebhookSubscriptionList,
	).WithSID("entity-read-access").When(organizationCondition),

	policy.Allow(
		ActionTrustCenterGet,
		ActionTrustCenterAccessGet, ActionTrustCenterAccessList,
		ActionTrustCenterDocumentAccessList,
		ActionTrustCenterFileGet, ActionTrustCenterFileList, ActionTrustCenterFileGetFileUrl,
		ActionTrustCenterReferenceList, ActionTrustCenterReferenceGetLogoUrl,
	).WithSID("trust-center-read-access").When(organizationCondition),

	policy.Allow(ActionCustomDomainGet).WithSID("custom-domain-read").When(organizationCondition),
	policy.Allow(ActionOrganizationContextGet).WithSID("organization-context-read").When(organizationCondition),
	policy.Allow(
		ActionDocumentVersionExportPDF, ActionDocumentVersionExportSignable, ActionDocumentVersionSign,
	).WithSID("document-signing").When(organizationCondition),

	policy.Allow(
		ActionProcessingActivityExport,
		ActionDataProtectionImpactAssessmentExport,
		ActionTransferImpactAssessmentExport,
	).WithSID("processing-activity-export").When(organizationCondition),
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

	policy.Allow(
		ActionVendorGet, ActionVendorList,
		ActionVendorContactGet, ActionVendorContactList,
		ActionVendorServiceGet, ActionVendorServiceList,
		ActionVendorComplianceReportGet, ActionVendorComplianceReportList,
		ActionVendorBusinessAssociateAgreementGet,
		ActionVendorDataPrivacyAgreementGet,
		ActionVendorRiskAssessmentList,
		ActionFrameworkGet, ActionFrameworkList,
		ActionControlGet, ActionControlList,
		ActionMeasureGet, ActionMeasureList,
		ActionEvidenceList,
		ActionDocumentGet, ActionDocumentList,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSignatureGet, ActionDocumentVersionSignatureList,
		ActionRiskGet, ActionRiskList,
		ActionAssetGet, ActionAssetList,
		ActionDatumGet, ActionDatumList,
		ActionAuditGet, ActionAuditList,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionNonconformityGet, ActionNonconformityList,
		ActionObligationGet, ActionObligationList,
		ActionContinualImprovementGet, ActionContinualImprovementList,
		ActionProcessingActivityGet, ActionProcessingActivityList,
		ActionDataProtectionImpactAssessmentGet,
		ActionTransferImpactAssessmentGet, ActionTransferImpactAssessmentList,
		ActionSnapshotGet, ActionSnapshotList,
		ActionMeetingGet, ActionMeetingList,
		ActionFileGet, ActionFileDownloadUrl,
	).WithSID("entity-read-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentVersionExportPDF, ActionDocumentVersionExportSignable, ActionDocumentVersionSign,
	).WithSID("document-signing").When(organizationCondition),
).WithDescription("Read-only probo access for auditors (excludes internal/employee content)")

// EmployeePolicy defines permissions for employee role.
var EmployeePolicy = policy.NewPolicy(
	"probo:employee",
	"Probo Employee",
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
	).WithSID("org-basic-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentGet, ActionDocumentList,
	).WithSID("document-signing-access").When(organizationCondition),

	policy.Allow(
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSign,
		ActionDocumentVersionExportSignable,
	).WithSID("document-version-signing").When(organizationCondition),
).WithDescription("Employee access - can sign documents and view internal content")

// ProboPolicySet returns the PolicySet for the probo service.
func ProboPolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", OwnerPolicy).
		AddRolePolicy("ADMIN", AdminPolicy).
		AddRolePolicy("VIEWER", ViewerPolicy).
		AddRolePolicy("AUDITOR", AuditorPolicy).
		AddRolePolicy("EMPLOYEE", EmployeePolicy)
}
