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

package oauth2scope

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"sync"

	"go.probo.inc/probo/pkg/coredata"
)

type Registry struct {
	mu            sync.RWMutex
	scopeActions  map[coredata.OAuth2Scope][]string
	invertedIndex map[string][]coredata.OAuth2Scope
}

func NewRegistry() *Registry {
	return &Registry{
		scopeActions: make(map[coredata.OAuth2Scope][]string),
	}
}

func (r *Registry) Register(mappings map[coredata.OAuth2Scope][]string) *Registry {
	r.mu.Lock()
	defer r.mu.Unlock()

	for scope, actions := range mappings {
		if len(actions) == 0 {
			continue
		}

		r.scopeActions[scope] = append(r.scopeActions[scope], actions...)
	}

	r.rebuildInvertedIndex()

	return r
}

func (r *Registry) RegisteredScopes() []coredata.OAuth2Scope {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return sortedScopes(slices.Collect(maps.Keys(r.scopeActions)))
}

func (r *Registry) Allows(tokenScopes coredata.OAuth2Scopes, action string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	grantingScopes, ok := r.invertedIndex[action]
	if !ok {
		return false
	}

	return slices.ContainsFunc(grantingScopes, tokenScopes.Contains)
}

func (r *Registry) ScopesForAction(action string) []coredata.OAuth2Scope {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return sortedScopes(r.invertedIndex[action])
}

func (r *Registry) ValidateScopes(scopes coredata.OAuth2Scopes) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, scope := range scopes {
		if _, ok := r.scopeActions[scope]; !ok {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

func (r *Registry) rebuildInvertedIndex() {
	invertedIndex := make(map[string][]coredata.OAuth2Scope, len(r.scopeActions)*4)

	for scope, actions := range r.scopeActions {
		for _, action := range actions {
			invertedIndex[action] = append(invertedIndex[action], scope)
		}
	}

	r.invertedIndex = invertedIndex
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
