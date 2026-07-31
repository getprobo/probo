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

package resourcetag

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

// FullAccessPolicy grants complete resource-tag access to organization owners
// and admins.
var FullAccessPolicy = policy.NewPolicy(
	"resourcetag:full-access",
	"Resource Tag Full Access",
	policy.Allow(
		ActionTagList,
		ActionTagGet,
		ActionTagCreate,
		ActionTagUpdate,
		ActionTagDelete,
		ActionAssignmentGet,
		ActionAssignmentAttach,
		ActionAssignmentDetach,
	).WithSID("resource-tag-full-access").When(organizationCondition),
).WithDescription("Full resource-tag access including catalog and assignments")

// ReadAccessPolicy grants read-only resource-tag access to viewers and auditors.
var ReadAccessPolicy = policy.NewPolicy(
	"resourcetag:read-access",
	"Resource Tag Read Access",
	policy.Allow(
		ActionTagList,
		ActionTagGet,
		ActionAssignmentGet,
	).WithSID("resource-tag-read-access").When(organizationCondition),
).WithDescription("Read-only resource-tag access")

// PolicySet returns the PolicySet for the resource-tag service.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy).
		AddRolePolicy("AUDITOR", ReadAccessPolicy)
}
