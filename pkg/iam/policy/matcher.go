// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
