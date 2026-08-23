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

package connector

import (
	"sort"
	"strings"
	"unicode"
)

// ParseScopeString splits an OAuth2 scope string into a sorted,
// deduplicated slice. Accepts both RFC 6749 §3.3 space-separated form
// (the standard) and GitHub's non-compliant comma-separated form. An
// empty or whitespace-only input returns an empty slice.
func ParseScopeString(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
	if len(fields) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(fields))

	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if _, ok := seen[f]; ok {
			continue
		}

		seen[f] = struct{}{}
		out = append(out, f)
	}

	sort.Strings(out)

	return out
}

// FormatScopeString joins scopes into the RFC 6749 §3.3 space-separated
// form. The output order is deterministic (sorted).
func FormatScopeString(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}

	sorted := make([]string, len(scopes))
	copy(sorted, scopes)
	sort.Strings(sorted)

	return strings.Join(sorted, " ")
}

// microsoftGraphScopePrefix is stripped when comparing scopes so that
// Microsoft's short permission names (User.Read.All) match the full
// resource URIs we request (https://graph.microsoft.com/User.Read.All).
const microsoftGraphScopePrefix = "https://graph.microsoft.com/"

// canonicalizeScope normalizes a scope string for equality checks.
func canonicalizeScope(scope string) string {
	return strings.TrimPrefix(scope, microsoftGraphScopePrefix)
}

// MissingScopes returns the sorted list of scopes present in required but
// absent from granted. Empty strings in either input are ignored. Scope
// comparison is canonicalized so Microsoft Graph short names and full
// resource URIs are treated as equivalent. The result uses the required
// scope strings as provided and never aliases any input.
func MissingScopes(required, granted []string) []string {
	grantedSet := make(map[string]struct{}, len(granted))
	for _, s := range granted {
		if s == "" {
			continue
		}

		grantedSet[canonicalizeScope(s)] = struct{}{}
	}

	missing := make([]string, 0)
	seenMissing := make(map[string]struct{})

	for _, s := range required {
		if s == "" {
			continue
		}

		key := canonicalizeScope(s)
		if _, ok := grantedSet[key]; ok {
			continue
		}

		if _, ok := seenMissing[key]; ok {
			continue
		}

		seenMissing[key] = struct{}{}

		missing = append(missing, s)
	}

	sort.Strings(missing)

	return missing
}

// GrantedScopes returns the scopes a stored connection's grant actually
// carries, for comparison against what a provider requires.
//
// Microsoft (and similar OIDC providers) omit offline_access from the token
// scope echo even when a refresh token was issued, so a refresh token is taken
// as proof of that grant. This applies to missing-scope checks only: it is
// never synthesized into Connection.Scopes(), which reconnect uses to build the
// next authorize request (Google rejects offline_access).
func GrantedScopes(c Connection) []string {
	if c == nil {
		return nil
	}

	granted := c.Scopes()

	if hasRefreshToken(c) {
		granted = UnionScopes(granted, []string{"offline_access"})
	}

	return granted
}

func hasRefreshToken(c Connection) bool {
	switch conn := c.(type) {
	case *OAuth2Connection:
		return conn.RefreshToken != ""
	case *SlackConnection:
		return conn.RefreshToken != ""
	}

	return false
}

// UnionScopes returns the sorted, deduplicated union of the given scope
// slices. Empty strings and empty slices are handled gracefully. The
// result is a fresh slice and never aliases any input.
func UnionScopes(scopeSets ...[]string) []string {
	seen := map[string]struct{}{}

	for _, set := range scopeSets {
		for _, s := range set {
			if s == "" {
				continue
			}

			seen[s] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}

	sort.Strings(out)

	return out
}
