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

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/iam/policy"
)

func TestPolicySet_AddAndMerge(t *testing.T) {
	// Create first policy set (simulating IAM service)
	iamPolicies := NewPolicySet().
		AddRolePolicy("OWNER", policy.NewPolicy("iam-owner", "IAM Owner", policy.Allow("iam:*"))).
		AddRolePolicy("ADMIN", policy.NewPolicy("iam-admin", "IAM Admin", policy.Allow("iam:read:*"))).
		AddIdentityScopedPolicy(policy.NewPolicy("iam-self", "IAM Self", policy.Allow("iam:identity:get")))

	// Create second policy set (simulating Documents service)
	docsPolicies := NewPolicySet().
		AddRolePolicy("OWNER", policy.NewPolicy("docs-owner", "Docs Owner", policy.Allow("docs:*"))).
		AddRolePolicy("VIEWER", policy.NewPolicy("docs-viewer", "Docs Viewer", policy.Allow("docs:read:*"))).
		AddIdentityScopedPolicy(policy.NewPolicy("docs-self", "Docs Self", policy.Allow("docs:own:*")))

	// Merge them
	combined := iamPolicies.Merge(docsPolicies)
	require.NotNil(t, combined, "combined policy set should not be nil")

	// Test OWNER has policies from both services
	ownerPolicies := combined.RolePolicies["OWNER"]
	require.Len(t, ownerPolicies, 2, "should have 2 OWNER policies")

	// Test ADMIN only has IAM policy
	adminPolicies := combined.RolePolicies["ADMIN"]
	require.Len(t, adminPolicies, 1, "should have 1 ADMIN policy")

	// Test VIEWER only has Docs policy
	viewerPolicies := combined.RolePolicies["VIEWER"]
	require.Len(t, viewerPolicies, 1, "should have 1 VIEWER policy")

	// Test self-manage policies from both services
	identityPolicies := combined.IdentityScopedPolicies
	require.Len(t, identityPolicies, 2, "should have 2 identity-scoped policies")
}

func TestIAMPolicySet(t *testing.T) {
	policySet := IAMPolicySet()
	require.NotNil(t, policySet, "IAMPolicySet should not return nil")

	// Should have policies for all standard roles
	roles := []string{"OWNER", "ADMIN", "VIEWER", "EMPLOYEE", "AUDITOR"}
	for _, role := range roles {
		policies := policySet.RolePolicies[role]
		assert.NotEmptyf(t, policies, "expected policies for role %s", role)
	}

	// Should have self-manage policies
	assert.NotEmpty(t, policySet.IdentityScopedPolicies, "expected identity-scoped policies")
}
