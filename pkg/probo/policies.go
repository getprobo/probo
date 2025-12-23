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

// OwnerPolicy defines permissions for organization owners.
var OwnerPolicy = policy.NewPolicy(
	"probo:owner",
	"Probo Owner",
	// Full access to all probo resources
	policy.Allow("core:*").WithSID("full-core-access"),
).WithDescription("Full probo access for organization owners")

// AdminPolicy defines permissions for organization admins.
var AdminPolicy = policy.NewPolicy(
	"probo:admin",
	"Probo Admin",
	// Full access to all probo resources (same as owner for core entities)
	policy.Allow("core:*").WithSID("full-core-access"),
).WithDescription("Probo admin access - can manage core entities")

// ViewerPolicy defines read-only permissions for organization viewers.
var ViewerPolicy = policy.NewPolicy(
	"probo:viewer",
	"Probo Viewer",
	// Organization read actions
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
		ActionOrganizationGetHorizontalLogoUrl,
	).WithSID("org-read-access"),

	// Entity read actions
	policy.Allow(
		ActionPeopleGet, ActionPeopleList,
		ActionVendorList,
		ActionFrameworkGet, ActionFrameworkList,
		ActionControlList,
		ActionMeasureGet, ActionMeasureList,
		ActionTaskGet, ActionTaskList,
		ActionEvidenceList,
		ActionDocumentGet, ActionDocumentList,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSignatureList,
		ActionRiskList,
		ActionAssetList,
		ActionDatumList,
		ActionAuditGet, ActionAuditList,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionNonconformityList,
		ActionObligationList,
		ActionContinualImprovementList,
		ActionProcessingActivityList,
		ActionSnapshotList,
		ActionMeetingList,
		ActionFileGet, ActionFileDownloadUrl,
		ActionSlackConnectionList,
	).WithSID("entity-read-access"),

	// TrustCenter read actions
	policy.Allow(
		ActionTrustCenterGet,
		ActionTrustCenterFileGet, ActionTrustCenterFileList,
	).WithSID("trust-center-read-access"),

	// CustomDomain read actions
	policy.Allow(ActionCustomDomainGet).WithSID("custom-domain-read"),

	// OrganizationContext read actions
	policy.Allow(ActionOrganizationContextGet).WithSID("organization-context-read"),

	// Document signing actions
	policy.Allow(
		ActionDocumentVersionExportPDF, ActionDocumentVersionExportSignable, ActionDocumentVersionSign,
	).WithSID("document-signing"),
).WithDescription("Read-only probo access for organization viewers")

// AuditorPolicy defines permissions for auditor role.
// Auditors have read access to non-employee content plus some specific auditor features.
var AuditorPolicy = policy.NewPolicy(
	"probo:auditor",
	"Probo Auditor",
	// Same as viewer but without employee-specific content
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
		ActionOrganizationGetHorizontalLogoUrl,
	).WithSID("org-read-access"),

	// Entity read access (same as viewer)
	policy.Allow(
		ActionPeopleGet, ActionPeopleList,
		ActionVendorList,
		ActionFrameworkGet, ActionFrameworkList,
		ActionControlList,
		ActionMeasureGet, ActionMeasureList,
		ActionEvidenceList,
		ActionDocumentGet, ActionDocumentList,
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSignatureList,
		ActionRiskList,
		ActionAssetList,
		ActionDatumList,
		ActionAuditGet, ActionAuditList,
		ActionReportGet, ActionReportGetReportUrl, ActionReportDownloadUrlGet,
		ActionNonconformityList,
		ActionObligationList,
		ActionContinualImprovementList,
		ActionProcessingActivityList,
		ActionSnapshotList,
		ActionFileGet, ActionFileDownloadUrl,
	).WithSID("entity-read-access"),

	// Document signing actions
	policy.Allow(
		ActionDocumentVersionExportPDF, ActionDocumentVersionExportSignable, ActionDocumentVersionSign,
	).WithSID("document-signing"),
).WithDescription("Read-only probo access for auditors (excludes internal/employee content)")

// EmployeePolicy defines permissions for employee role.
// Employees have access to internal documents and some limited read access.
var EmployeePolicy = policy.NewPolicy(
	"probo:employee",
	"Probo Employee",
	// Basic organization access
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationGetLogoUrl,
	).WithSID("org-basic-access"),

	// Document signing access
	policy.Allow(
		ActionDocumentGet, ActionDocumentList,
	).WithSID("document-signing-access"),

	// Document version signing
	policy.Allow(
		ActionDocumentVersionGet, ActionDocumentVersionList,
		ActionDocumentVersionSign,
		ActionDocumentVersionExportSignable,
	).WithSID("document-version-signing"),
).WithDescription("Employee access - can sign documents and view internal content")

// ProboPolicySet returns the PolicySet for the probo service.
// This is registered with the IAM Authorizer when probo.Service is created.
func ProboPolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", OwnerPolicy).
		AddRolePolicy("ADMIN", AdminPolicy).
		AddRolePolicy("VIEWER", ViewerPolicy).
		AddRolePolicy("AUDITOR", AuditorPolicy).
		AddRolePolicy("EMPLOYEE", EmployeePolicy)
}
