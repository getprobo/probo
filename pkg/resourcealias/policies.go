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
