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

package resourcealias

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

// FullAccessPolicy grants complete resource-alias access to organization owners
// and admins.
var FullAccessPolicy = policy.NewPolicy(
	"resourcealias:full-access",
	"Resource Alias Full Access",
	policy.Allow(
		ActionAliasGet,
		ActionAliasSet,
		ActionAliasRemove,
	).WithSID("resource-alias-full-access").When(organizationCondition),
).WithDescription("Full resource-alias access including set and remove")

// ReadAccessPolicy grants read-only resource-alias access to viewers and auditors.
var ReadAccessPolicy = policy.NewPolicy(
	"resourcealias:read-access",
	"Resource Alias Read Access",
	policy.Allow(
		ActionAliasGet,
	).WithSID("resource-alias-read-access").When(organizationCondition),
).WithDescription("Read-only resource-alias access")

// PolicySet returns the PolicySet for the resource-alias service. It is owned by
// this package and registered into the authorizer at composition time so the
// resource-alias authorization rules live alongside the resource-alias domain
// logic instead of in the core probo policy set.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy).
		AddRolePolicy("AUDITOR", ReadAccessPolicy)
}
