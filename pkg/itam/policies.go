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

package itam

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

// ViewerPolicy grants read-only access to ITAM entities for organization
// viewers. Owners and admins are already covered by the probo OWNER/ADMIN
// wildcards.
var ViewerPolicy = policy.NewPolicy(
	"itam:viewer",
	"ITAM Viewer",
	policy.Allow(
		ActionDeviceGet, ActionDeviceList,
		ActionDevicePostureList,
		ActionDeviceEnrollmentTokenGet, ActionDeviceEnrollmentTokenList,
	).WithSID("itam-read-access").When(organizationCondition),
).WithDescription("Read-only ITAM access for organization viewers")

// ITAMPolicySet returns the PolicySet for the ITAM service.
func ITAMPolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("VIEWER", ViewerPolicy)
}
