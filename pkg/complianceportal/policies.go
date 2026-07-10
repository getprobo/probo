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

package complianceportal

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

// FullAccessPolicy grants organization owners and admins complete access to
// every compliance portal capability, including custom domains, portal
// configuration, access grants, files, references, frameworks, external URLs
// and mailing lists.
//
// The managed probopage subdomain is a system-owned resource and can never be
// deleted, so an explicit deny (which takes precedence over any allow) blocks
// deletion of managed domains for every role.
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

// ViewerPolicy grants organization viewers read-only access to the compliance
// portal.
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
		ActionComplianceFrameworkList,
	).WithSID("compliance-portal-read-access").When(organizationCondition),
).WithDescription("Read-only compliance portal access for organization viewers")

// PolicySet returns the PolicySet for the compliance portal service.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ViewerPolicy)
}
