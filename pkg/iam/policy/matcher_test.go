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

import (
	"testing"
)

func TestActionMatcher_Matches(t *testing.T) {
	m := NewActionMatcher()

	tests := []struct {
		name    string
		pattern string
		target  string
		want    bool
	}{
		// Exact matches
		{
			name:    "exact match",
			pattern: "iam:identity:get",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "exact no match - different operation",
			pattern: "iam:identity:get",
			target:  "iam:identity:update",
			want:    false,
		},
		{
			name:    "exact no match - different resource",
			pattern: "iam:identity:get",
			target:  "iam:organization:get",
			want:    false,
		},
		{
			name:    "exact no match - different service",
			pattern: "iam:identity:get",
			target:  "documents:identity:get",
			want:    false,
		},

		// Full wildcard
		{
			name:    "full wildcard",
			pattern: "*",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "full wildcard matches any action",
			pattern: "*",
			target:  "documents:document:delete",
			want:    true,
		},

		// Operation wildcard
		{
			name:    "operation wildcard",
			pattern: "iam:identity:*",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "operation wildcard matches update",
			pattern: "iam:identity:*",
			target:  "iam:identity:update",
			want:    true,
		},
		{
			name:    "operation wildcard no match - different resource",
			pattern: "iam:identity:*",
			target:  "iam:organization:get",
			want:    false,
		},

		// Resource wildcard
		{
			name:    "resource wildcard",
			pattern: "iam:*:get",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "resource wildcard matches organization",
			pattern: "iam:*:get",
			target:  "iam:organization:get",
			want:    true,
		},
		{
			name:    "resource wildcard no match - different operation",
			pattern: "iam:*:get",
			target:  "iam:identity:update",
			want:    false,
		},

		// Service wildcard
		{
			name:    "service wildcard",
			pattern: "*:identity:get",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "service wildcard matches documents",
			pattern: "*:document:read",
			target:  "documents:document:read",
			want:    true,
		},

		// Multiple wildcards
		{
			name:    "service and operation wildcard",
			pattern: "*:identity:*",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "resource and operation wildcard",
			pattern: "iam:*:*",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "resource and operation wildcard matches any iam action",
			pattern: "iam:*:*",
			target:  "iam:organization:delete",
			want:    true,
		},
		{
			name:    "resource and operation wildcard no match - different service",
			pattern: "iam:*:*",
			target:  "documents:document:read",
			want:    false,
		},
		{
			name:    "all wildcards",
			pattern: "*:*:*",
			target:  "anything:goes:here",
			want:    true,
		},

		// Two-part pattern (service:*)
		{
			name:    "two-part pattern service wildcard",
			pattern: "iam:*",
			target:  "iam:identity:get",
			want:    true,
		},
		{
			name:    "two-part pattern service wildcard no match",
			pattern: "iam:*",
			target:  "documents:document:read",
			want:    false,
		},

		// Invalid targets
		{
			name:    "invalid target - too few parts",
			pattern: "iam:identity:get",
			target:  "iam:identity",
			want:    false,
		},
		{
			name:    "invalid target - single part",
			pattern: "iam:identity:get",
			target:  "iam",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.Matches(tt.pattern, tt.target)
			if got != tt.want {
				t.Errorf("Matches(%q, %q) = %v, want %v", tt.pattern, tt.target, got, tt.want)
			}
		})
	}
}

func TestActionMatcher_MatchesAny(t *testing.T) {
	m := NewActionMatcher()

	tests := []struct {
		name     string
		patterns []string
		target   string
		want     bool
	}{
		{
			name:     "matches first pattern",
			patterns: []string{"iam:identity:get", "iam:identity:update"},
			target:   "iam:identity:get",
			want:     true,
		},
		{
			name:     "matches second pattern",
			patterns: []string{"iam:identity:get", "iam:identity:update"},
			target:   "iam:identity:update",
			want:     true,
		},
		{
			name:     "no match",
			patterns: []string{"iam:identity:get", "iam:identity:update"},
			target:   "iam:identity:delete",
			want:     false,
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			target:   "iam:identity:get",
			want:     false,
		},
		{
			name:     "wildcard in patterns",
			patterns: []string{"iam:*:get", "documents:*:read"},
			target:   "iam:organization:get",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MatchesAny(tt.patterns, tt.target)
			if got != tt.want {
				t.Errorf("MatchesAny(%v, %q) = %v, want %v", tt.patterns, tt.target, got, tt.want)
			}
		})
	}
}
