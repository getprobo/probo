// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

// ScopesCover reports whether `granted` already contains every scope in
// `required`. An empty `required` set is trivially covered.
func ScopesCover(granted, required []string) bool {
	if len(required) == 0 {
		return true
	}
	grantedSet := make(map[string]struct{}, len(granted))
	for _, s := range granted {
		grantedSet[s] = struct{}{}
	}
	for _, r := range required {
		if _, ok := grantedSet[r]; !ok {
			return false
		}
	}
	return true
}
