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

func (r *Registry) AllWriteScopes() []coredata.OAuth2Scope {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writeScopes := make([]coredata.OAuth2Scope, 0, len(r.scopeActions))
	for scope := range r.scopeActions {
		if !scope.IsRead() {
			writeScopes = append(writeScopes, scope)
		}
	}

	return sortedScopes(writeScopes)
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
