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

package iam

import "go.probo.inc/probo/pkg/iam/policy"

// PolicySet holds organization-scoped (role) policies and identity-scoped policies.
// Services create their own PolicySet and combine them when creating the Authorizer.
type PolicySet struct {
	// RolePolicies maps role names to policies.
	RolePolicies map[string][]*policy.Policy

	// IdentityScopedPolicies are applied to all authenticated users, independent of organization membership.
	IdentityScopedPolicies []*policy.Policy
}

// NewPolicySet creates an empty PolicySet.
func NewPolicySet() *PolicySet {
	return &PolicySet{
		RolePolicies:           make(map[string][]*policy.Policy),
		IdentityScopedPolicies: make([]*policy.Policy, 0),
	}
}

// AddRolePolicy adds a policy for a specific role.
func (ps *PolicySet) AddRolePolicy(role string, policies ...*policy.Policy) *PolicySet {
	ps.RolePolicies[role] = append(ps.RolePolicies[role], policies...)
	return ps
}

// AddIdentityScopedPolicy adds policies applied to all authenticated users (identity-scoped).
func (ps *PolicySet) AddIdentityScopedPolicy(policies ...*policy.Policy) *PolicySet {
	ps.IdentityScopedPolicies = append(ps.IdentityScopedPolicies, policies...)
	return ps
}

// Merge combines another PolicySet into this one.
func (ps *PolicySet) Merge(other *PolicySet) *PolicySet {
	for role, policies := range other.RolePolicies {
		ps.RolePolicies[role] = append(ps.RolePolicies[role], policies...)
	}

	ps.IdentityScopedPolicies = append(ps.IdentityScopedPolicies, other.IdentityScopedPolicies...)

	return ps
}

func IAMPolicySet() *PolicySet {
	return NewPolicySet().
		AddRolePolicy("OWNER", IAMOwnerPolicy).
		AddRolePolicy("ADMIN", IAMAdminPolicy).
		AddRolePolicy("VIEWER", IAMViewerPolicy).
		AddRolePolicy("EMPLOYEE", IAMViewerPolicy).
		AddRolePolicy("AUDITOR", IAMViewerPolicy).
		AddIdentityScopedPolicy(
			IAMSelfManageIdentityPolicy,
			IAMSelfManageSessionPolicy,
			IAMSelfManageInvitationPolicy,
			IAMSelfManageProfilePolicy,
			IAMSelfManageMembershipPolicy,
			IAMSelfManagePersonalAPIKeyPolicy,
			IAMSelfManageOAuth2AccessTokenPolicy,
			IAMSelfManageOAuth2ConsentPolicy,
		)
}
