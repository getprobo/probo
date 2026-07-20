// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package management

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

var FullAccessPolicy = policy.NewPolicy(
	"compliance-portal:full-access",
	"Compliance Portal Full Access",
	policy.Allow("compliance-portal:*").
		WithSID("compliance-portal-full-access").
		When(organizationCondition),
	policy.Deny(ActionCustomDomainDelete).
		WithSID("custom-domain-managed-no-delete").
		When(policy.Equals("resource.managed", "true")),
).WithDescription("Full compliance portal access for organization owners and admins")

var ViewerPolicy = policy.NewPolicy(
	"compliance-portal:viewer",
	"Compliance Portal Viewer",
	policy.Allow(
		ActionCustomDomainGet,
		ActionCompliancePortalGet,
		ActionCompliancePortalAccessGet, ActionCompliancePortalAccessList,
		ActionCompliancePortalDocumentAccessList,
		ActionCompliancePortalFileGet, ActionCompliancePortalFileList, ActionCompliancePortalFileGetFileUrl,
		ActionCompliancePortalReferenceList, ActionCompliancePortalReferenceGetLogoUrl,
		ActionCompliancePortalCommitmentGroupList, ActionCompliancePortalCommitmentList,
		ActionComplianceFrameworkList,
		ActionComplianceCustomLinkList,
	).WithSID("compliance-portal-read-access").When(organizationCondition),
).WithDescription("Read-only compliance portal access for organization viewers")

func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ViewerPolicy)
}
