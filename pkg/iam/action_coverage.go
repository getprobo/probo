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

package iam

import (
	"fmt"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/iam/policy"
)

// MustCoverDeclaredActions panics when any declared action is not matched by the
// registered policy set. Call after all policy sets are registered.
func (a *Authorizer) MustCoverDeclaredActions(declared []string) {
	uncovered := UncoveredDeclaredActions(a.policySet, declared)
	if len(uncovered) == 0 {
		return
	}

	panic(fmt.Errorf("uncovered authz actions: %s", strings.Join(uncovered, ", ")))
}

// UncoveredDeclaredActions returns declared actions with no matching policy pattern.
func UncoveredDeclaredActions(ps *PolicySet, declared []string) []string {
	patterns := policyActionPatterns(ps)
	matcher := policy.NewActionMatcher()

	uncovered := make([]string, 0)

	for _, action := range declared {
		covered := false

		for _, pattern := range patterns {
			if matcher.Matches(pattern, action) {
				covered = true
				break
			}
		}

		if !covered {
			uncovered = append(uncovered, action)
		}
	}

	slices.Sort(uncovered)

	return uncovered
}

func policyActionPatterns(ps *PolicySet) []string {
	seen := make(map[string]struct{})
	patterns := make([]string, 0)

	for _, p := range uniquePolicies(ps) {
		for _, stmt := range p.Statements {
			for _, action := range stmt.Actions {
				if _, ok := seen[action]; ok {
					continue
				}

				seen[action] = struct{}{}
				patterns = append(patterns, action)
			}
		}
	}

	slices.Sort(patterns)

	return patterns
}

func uniquePolicies(ps *PolicySet) []*policy.Policy {
	seen := make(map[string]struct{})
	policies := make([]*policy.Policy, 0)

	for _, p := range ps.IdentityScopedPolicies {
		if _, ok := seen[p.ID]; ok {
			continue
		}

		seen[p.ID] = struct{}{}
		policies = append(policies, p)
	}

	for _, rolePolicies := range ps.RolePolicies {
		for _, p := range rolePolicies {
			if _, ok := seen[p.ID]; ok {
				continue
			}

			seen[p.ID] = struct{}{}
			policies = append(policies, p)
		}
	}

	return policies
}
