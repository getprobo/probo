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

package itam

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var (
	organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")
	ownerCondition        = policy.Equals("principal.id", "resource.owner_id")
)

// FullAccessPolicy grants complete ITAM access to organization owners and
// admins.
var FullAccessPolicy = policy.NewPolicy(
	"itam:full-access",
	"ITAM Full Access",
	policy.Allow(
		ActionDeviceList, ActionEmployeeDeviceList, ActionDeviceGet, ActionDeviceCreate,
		ActionDeviceEnroll, ActionDeviceRevoke, ActionDeviceAssignOwner,
		ActionDevicePostureList,
	).WithSID("itam-full-access").When(organizationCondition),
).WithDescription("Full ITAM access for organization owners and admins")

// ViewerPolicy grants read-only access to ITAM entities for organization
// viewers.
var ViewerPolicy = policy.NewPolicy(
	"itam:viewer",
	"ITAM Viewer",
	policy.Allow(
		ActionDeviceGet, ActionDeviceList,
		ActionDevicePostureList,
	).WithSID("itam-read-access").When(organizationCondition),
).WithDescription("Read-only ITAM access for organization viewers")

// EmployeePolicy grants self-enrollment access to organization employees.
var EmployeePolicy = policy.NewPolicy(
	"itam:employee",
	"ITAM Employee",
	policy.Allow(ActionDeviceEnroll).
		WithSID("itam-employee-enroll-device").
		When(organizationCondition),
	policy.Allow(ActionDeviceGet).
		WithSID("itam-employee-get-own-device").
		When(organizationCondition, ownerCondition),
	policy.Allow(ActionEmployeeDeviceList).
		WithSID("itam-employee-device-list").
		When(organizationCondition),
).WithDescription("Self-enrollment: enroll own device and read own enrolled devices")

// ITAMPolicySet returns the PolicySet for the ITAM service.
func ITAMPolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ViewerPolicy).
		AddRolePolicy("EMPLOYEE", EmployeePolicy)
}
