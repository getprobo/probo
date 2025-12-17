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

package policy

import "strings"

// ActionMatcher handles wildcard matching for actions.
// Supported patterns:
//   - Exact match: "documents:document:read"
//   - Single wildcard: "documents:document:*" (any operation)
//   - Service wildcard: "documents:*:*" (any resource/operation in service)
//   - Full wildcard: "*" (matches everything)
type ActionMatcher struct{}

// NewActionMatcher creates a new action matcher.
func NewActionMatcher() *ActionMatcher {
	return &ActionMatcher{}
}

// Matches checks if a pattern matches a target action.
// Pattern can contain wildcards (*), target should be a concrete action.
func (m *ActionMatcher) Matches(pattern, target string) bool {
	// Full wildcard
	if pattern == "*" {
		return true
	}

	patternParts := strings.Split(pattern, ":")
	targetParts := strings.Split(target, ":")

	// Both should have the same number of parts (3) for service:resource:operation
	if len(targetParts) != 3 {
		return false
	}

	// Pattern can have 1-3 parts
	switch len(patternParts) {
	case 1:
		// Single part pattern (should be "*" which is handled above)
		return false

	case 2:
		// Two parts: "service:*" means "service:*:*"
		if patternParts[1] == "*" {
			return patternParts[0] == targetParts[0] || patternParts[0] == "*"
		}
		return false

	case 3:
		// Full pattern: service:resource:operation
		return m.matchPart(patternParts[0], targetParts[0]) &&
			m.matchPart(patternParts[1], targetParts[1]) &&
			m.matchPart(patternParts[2], targetParts[2])

	default:
		return false
	}
}

// matchPart checks if a single part matches (exact or wildcard).
func (m *ActionMatcher) matchPart(pattern, target string) bool {
	if pattern == "*" {
		return true
	}
	return pattern == target
}

// MatchesAny checks if any of the patterns match the target action.
func (m *ActionMatcher) MatchesAny(patterns []string, target string) bool {
	for _, pattern := range patterns {
		if m.Matches(pattern, target) {
			return true
		}
	}
	return false
}

