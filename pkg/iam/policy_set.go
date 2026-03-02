// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
		)
}
