// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package tag

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

// FullAccessPolicy grants complete tag access to organization owners
// and admins.
var FullAccessPolicy = policy.NewPolicy(
	"tag:full-access",
	"Tag Full Access",
	policy.Allow(
		ActionList,
		ActionGet,
		ActionCreate,
		ActionUpdate,
		ActionDelete,
		ActionAssignmentGet,
		ActionAssignmentAttach,
		ActionAssignmentDetach,
	).WithSID("tag-full-access").When(organizationCondition),
).WithDescription("Full tag access including catalog and assignments")

// ReadAccessPolicy grants read-only tag access to viewers and auditors.
var ReadAccessPolicy = policy.NewPolicy(
	"tag:read-access",
	"Tag Read Access",
	policy.Allow(
		ActionList,
		ActionGet,
		ActionAssignmentGet,
	).WithSID("tag-read-access").When(organizationCondition),
).WithDescription("Read-only tag access")

// PolicySet returns the PolicySet for the tag service.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy).
		AddRolePolicy("AUDITOR", ReadAccessPolicy)
}
