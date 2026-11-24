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

package accessreview

import (
	"strings"

	"go.probo.inc/probo/pkg/accessreview/drivers"
)

const (
	// DefaultResourceKey is the roster-path resource for a connected system /
	// selected org until drivers emit native grants.
	DefaultResourceKey = "system"

	// DefaultResourceKind is the kind stored on the default roster resource.
	DefaultResourceKind = "system"

	// ImplicitMemberEntitlementKey is used when an account has no usable roles.
	ImplicitMemberEntitlementKey = "member"

	// ImplicitMemberEntitlementDisplayName is the display label for the
	// implicit membership grant.
	ImplicitMemberEntitlementDisplayName = "Member"
)

// GrantTriple is one principal × resource × entitlement binding derived from
// a roster AccountRecord. Resource display names are supplied at upsert time
// from the access-review source, not from the account record.
type GrantTriple struct {
	PrincipalKey           string
	ResourceKey            string
	ResourceKind           string
	EntitlementKey         string
	EntitlementDisplayName string
}

// GrantTriplesFromAccount normalises one roster account into grant triples.
// The default resource is the connected system. Each non-empty role becomes
// an entitlement; when roles is empty (or only whitespace), a single implicit
// member entitlement is emitted. Duplicate entitlement keys after trim keep
// the first display name.
//
// Returns ErrMissingPrincipalKey when both email and external ID are empty
// after trim — an empty principal key would collide inventory upserts.
func GrantTriplesFromAccount(account drivers.AccountRecord) ([]GrantTriple, error) {
	principalKey := normalizeAccountKey(account.Email, account.ExternalID)
	if principalKey == "" {
		return nil, ErrMissingPrincipalKey
	}

	entitlements := rosterEntitlements(account.Roles)

	triples := make([]GrantTriple, 0, len(entitlements))
	for _, entitlement := range entitlements {
		triples = append(
			triples,
			GrantTriple{
				PrincipalKey:           principalKey,
				ResourceKey:            DefaultResourceKey,
				ResourceKind:           DefaultResourceKind,
				EntitlementKey:         entitlement.key,
				EntitlementDisplayName: entitlement.displayName,
			},
		)
	}

	return triples, nil
}

// GrantTriplesFromAccounts normalises a roster into grant triples, preserving
// account order and per-account entitlement order. Accounts with no usable
// email or external ID are skipped so they never emit an empty principal key.
func GrantTriplesFromAccounts(accounts []drivers.AccountRecord) []GrantTriple {
	triples := make([]GrantTriple, 0, len(accounts))
	for _, account := range accounts {
		accountTriples, err := GrantTriplesFromAccount(account)
		if err != nil {
			continue
		}

		triples = append(triples, accountTriples...)
	}

	return triples
}

type rosterEntitlement struct {
	key         string
	displayName string
}

func rosterEntitlements(roles []string) []rosterEntitlement {
	seen := make(map[string]struct{}, len(roles))
	entitlements := make([]rosterEntitlement, 0, len(roles))

	for _, role := range roles {
		key := strings.TrimSpace(role)
		if key == "" {
			continue
		}

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		entitlements = append(
			entitlements,
			rosterEntitlement{
				key:         key,
				displayName: key,
			},
		)
	}

	if len(entitlements) == 0 {
		return []rosterEntitlement{
			{
				key:         ImplicitMemberEntitlementKey,
				displayName: ImplicitMemberEntitlementDisplayName,
			},
		}
	}

	return entitlements
}
