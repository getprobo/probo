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
	"cmp"
	"maps"
	"slices"

	"go.probo.inc/probo/pkg/coredata"
)

// ScopeSet holds OAuth2 scope to IAM action mappings. Services create their
// own ScopeSet and register it on the Authorizer at composition time.
type ScopeSet struct {
	scopeActions map[coredata.OAuth2Scope][]Action
	actionScopes map[Action][]coredata.OAuth2Scope
}

// NewScopeSet creates an empty ScopeSet.
func NewScopeSet() *ScopeSet {
	return &ScopeSet{
		scopeActions: make(map[coredata.OAuth2Scope][]Action),
	}
}

// CreateScopeSet creates a ScopeSet from scope-to-action mappings. Entries with
// no actions are skipped.
func CreateScopeSet(mappings map[coredata.OAuth2Scope][]Action) *ScopeSet {
	s := NewScopeSet()

	for scope, actions := range mappings {
		if len(actions) == 0 {
			continue
		}

		s.scopeActions[scope] = append(s.scopeActions[scope], actions...)
	}

	s.rebuildActionScopes()

	return s
}

// Merge combines another ScopeSet into this one.
func (s *ScopeSet) Merge(other *ScopeSet) *ScopeSet {
	for scope, actions := range other.scopeActions {
		s.scopeActions[scope] = append(s.scopeActions[scope], actions...)
	}

	s.rebuildActionScopes()

	return s
}

// APIScopes returns every registered OAuth2 API scope in this set.
func (s *ScopeSet) APIScopes() []coredata.OAuth2Scope {
	return sortedScopes(slices.Collect(maps.Keys(s.scopeActions)))
}

// Allows reports whether tokenScopes authorize action.
func (s *ScopeSet) Allows(tokenScopes coredata.OAuth2Scopes, action Action) bool {
	grantingScopes, ok := s.actionScopes[action]
	if !ok {
		return false
	}

	return slices.ContainsFunc(grantingScopes, tokenScopes.Contains)
}

func (s *ScopeSet) rebuildActionScopes() {
	actionScopes := make(map[Action][]coredata.OAuth2Scope, len(s.scopeActions)*4)

	for scope, actions := range s.scopeActions {
		for _, action := range actions {
			actionScopes[action] = append(actionScopes[action], scope)
		}
	}

	s.actionScopes = actionScopes
}

func sortedScopes(scopes []coredata.OAuth2Scope) []coredata.OAuth2Scope {
	sorted := slices.Clone(scopes)
	slices.SortFunc(sorted, func(a, b coredata.OAuth2Scope) int {
		return cmp.Compare(string(a), string(b))
	})

	return sorted
}
