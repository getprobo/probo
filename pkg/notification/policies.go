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

package notification

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var (
	organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")
)

var actions = []string{
	ActionNotificationGet,
	ActionNotificationList,
	ActionNotificationUpdate,
}

// AccessPolicy grants inbox read and mark-read access. Owner and admin
// already match core:* on the probo role policy; this set keeps notification
// authorization next to the domain.
var AccessPolicy = policy.NewPolicy(
	"notification:access",
	"Notification Access",
	policy.Allow(actions...).
		WithSID("notification-access").
		When(organizationCondition),
).WithDescription("Read and mark-read access to the membership inbox")

// PolicySet returns the PolicySet for the notification service. It is owned
// by this package and registered into the authorizer at composition time so
// notification authorization rules live alongside the domain logic instead of
// in the core probo policy set.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", AccessPolicy).
		AddRolePolicy("ADMIN", AccessPolicy).
		AddRolePolicy("VIEWER", AccessPolicy).
		AddRolePolicy("AUDITOR", AccessPolicy).
		AddRolePolicy("EMPLOYEE", AccessPolicy).
		AddRolePolicy("COMPLIANCE_PORTAL_MANAGER", AccessPolicy).
		AddRolePolicy("COMPLIANCE_PORTAL_ACCESS_MANAGER", AccessPolicy)
}
