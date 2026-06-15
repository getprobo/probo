// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package iam_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/probo"
)

func TestUncoveredDeclaredActions_ProductionPolicySets(t *testing.T) {
	t.Parallel()

	policySet := iam.NewPolicySet()
	policySet.Merge(iam.IAMPolicySet())
	policySet.Merge(probo.ProboPolicySet())
	policySet.Merge(agentrun.PolicySet())

	declared := slices.Concat(
		probo.DeclaredActions(),
		iam.DeclaredActions(),
		agentrun.DeclaredActions(),
	)

	uncovered := iam.UncoveredDeclaredActions(policySet, declared)
	require.Empty(
		t,
		uncovered,
		"declared actions must be covered by registered policies: %v",
		uncovered,
	)
}

func TestMustCoverDeclaredActions_PanicsOnUncoveredAction(t *testing.T) {
	t.Parallel()

	authorizer := iam.NewAuthorizer(nil, nil)
	authorizer.RegisterPolicySet(
		iam.NewPolicySet().AddRolePolicy(
			"OWNER",
			policy.NewPolicy(
				"test-policy",
				"Test Policy",
				policy.Allow("core:example:get").WithSID("allow-example-get"),
			),
		),
	)

	assert.Panics(t, func() {
		authorizer.MustCoverDeclaredActions([]string{"core:example:delete"})
	})
}

func TestMustCoverDeclaredActions_AllowsCoveredAction(t *testing.T) {
	t.Parallel()

	authorizer := iam.NewAuthorizer(nil, nil)
	authorizer.RegisterPolicySet(
		iam.NewPolicySet().AddRolePolicy(
			"OWNER",
			policy.NewPolicy(
				"test-policy",
				"Test Policy",
				policy.Allow("core:example:get").WithSID("allow-example-get"),
			),
		),
	)

	assert.NotPanics(t, func() {
		authorizer.MustCoverDeclaredActions([]string{"core:example:get"})
	})
}

func TestMustCoverDeclaredActions_AllowsWildcardCoverage(t *testing.T) {
	t.Parallel()

	authorizer := iam.NewAuthorizer(nil, nil)
	authorizer.RegisterPolicySet(
		iam.NewPolicySet().AddRolePolicy(
			"OWNER",
			policy.NewPolicy(
				"test-policy",
				"Test Policy",
				policy.Allow("core:*").WithSID("allow-core"),
			),
		),
	)

	assert.NotPanics(t, func() {
		authorizer.MustCoverDeclaredActions([]string{"core:example:delete"})
	})
}
