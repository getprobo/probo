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

package scopeset

import (
	"cmp"
	"maps"
	"slices"
	"sync"

	"go.probo.inc/probo/pkg/coredata"
)

type ScopeSet struct {
	mu           sync.RWMutex
	scopeActions map[coredata.OAuth2Scope][]string
	actionScopes map[string][]coredata.OAuth2Scope
}

func New() *ScopeSet {
	return &ScopeSet{
		scopeActions: make(map[coredata.OAuth2Scope][]string),
	}
}

func (s *ScopeSet) Register(mappings map[coredata.OAuth2Scope][]string) *ScopeSet {
	s.mu.Lock()
	defer s.mu.Unlock()

	for scope, actions := range mappings {
		if len(actions) == 0 {
			continue
		}

		s.scopeActions[scope] = append(s.scopeActions[scope], actions...)
	}

	s.rebuildActionScopes()

	return s
}

func (s *ScopeSet) APIScopes() []coredata.OAuth2Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return sortedScopes(slices.Collect(maps.Keys(s.scopeActions)))
}

func (s *ScopeSet) Allows(tokenScopes coredata.OAuth2Scopes, action string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	grantingScopes, ok := s.actionScopes[action]
	if !ok {
		return false
	}

	return slices.ContainsFunc(grantingScopes, tokenScopes.Contains)
}

func (s *ScopeSet) rebuildActionScopes() {
	actionScopes := make(map[string][]coredata.OAuth2Scope, len(s.scopeActions)*4)

	for scope, actions := range s.scopeActions {
		for _, action := range actions {
			actionScopes[action] = append(actionScopes[action], scope)
		}
	}

	s.actionScopes = actionScopes
}

func sortedScopes(scopes []coredata.OAuth2Scope) []coredata.OAuth2Scope {
	sorted := slices.Clone(scopes)
	slices.SortFunc(
		sorted,
		func(a, b coredata.OAuth2Scope) int {
			return cmp.Compare(string(a), string(b))
		},
	)

	return sorted
}
