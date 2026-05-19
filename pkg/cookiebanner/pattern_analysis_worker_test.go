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

package cookiebanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestLooksVariable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "long mixed alphanumeric",
			token:    "XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN",
			expected: true,
		},
		{
			name:     "short mixed alphanumeric below threshold",
			token:    "abc123",
			expected: false,
		},
		{
			name:     "exactly 8 chars mixed",
			token:    "a1b2c3d4",
			expected: true,
		},
		{
			name:     "long hex string 16 chars",
			token:    "a1b2c3d4e5f60718",
			expected: true,
		},
		{
			name:     "short hex string below threshold",
			token:    "abcdef12",
			expected: true,
		},
		{
			name:     "pure letters not variable",
			token:    "posthog",
			expected: false,
		},
		{
			name:     "long pure letters not variable",
			token:    "authentication",
			expected: false,
		},
		{
			name:     "brand name with digit",
			token:    "auth0",
			expected: false,
		},
		{
			name:     "UUID shape",
			token:    "550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "8 digit number",
			token:    "12345678",
			expected: true,
		},
		{
			name:     "short digit number",
			token:    "12345",
			expected: false,
		},
		{
			name:     "empty string",
			token:    "",
			expected: false,
		},
		{
			name:     "single char",
			token:    "x",
			expected: false,
		},
		{
			name:     "all uppercase letters",
			token:    "MEASUREMENT",
			expected: false,
		},
		{
			name:     "short brand c15t",
			token:    "c15t",
			expected: false,
		},
		{
			name:     "GA measurement ID style",
			token:    "G-1234ABCDEF",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, looksVariable(tt.token))
			},
		)
	}
}

func TestIsUUIDShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid UUID lowercase",
			input:    "550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "valid UUID uppercase",
			input:    "550E8400-E29B-41D4-A716-446655440000",
			expected: true,
		},
		{
			name:     "wrong length",
			input:    "550e8400-e29b-41d4-a716",
			expected: false,
		},
		{
			name:     "no dashes",
			input:    "550e8400e29b41d4a716446655440000xxxx",
			expected: false,
		},
		{
			name:     "dashes in wrong positions",
			input:    "550e840-0e29b-41d4-a716-446655440000",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, isUUIDShape(tt.input))
			},
		)
	}
}

func TestHeuristicTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		template string
		changed  bool
	}{
		{
			name:     "posthog hash in middle",
			input:    "ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_window_id",
			template: "ph_phc_*_window_id",
			changed:  true,
		},
		{
			name:     "no variable tokens",
			input:    "probo_consent_given",
			template: "",
			changed:  false,
		},
		{
			name:     "trailing hash",
			input:    "auth0_session_a1b2c3d4e5f6",
			template: "auth0_session_*",
			changed:  true,
		},
		{
			name:     "no separator with variable token",
			input:    "a1b2c3d4e5f6g7h8",
			template: "",
			changed:  false,
		},
		{
			name:     "no separator without variable token",
			input:    "PHPSESSID",
			template: "",
			changed:  false,
		},
		{
			name:     "consecutive variable tokens collapse to single star",
			input:    "ph_a1b2c3d4_e5f6g7h8_window",
			template: "ph_*_window",
			changed:  true,
		},
		{
			name:     "UUID token replaced",
			input:    "session_550e8400-e29b-41d4-a716-446655440000_data",
			template: "session_*_data",
			changed:  true,
		},
		{
			name:     "leading underscore with hash",
			input:    "_ga_G1234ABCDEF",
			template: "_ga_*",
			changed:  true,
		},
		{
			name:     "dash separator with hash",
			input:    "c15t-consent-a1b2c3d4e5f6",
			template: "c15t-consent-*",
			changed:  true,
		},
		{
			name:    "leading underscores with dash not variable",
			input:   "__Secure-1PSID",
			changed: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				tmpl, changed := heuristicTemplate(tt.input)
				assert.Equal(t, tt.changed, changed)

				if changed {
					assert.Equal(t, tt.template, tmpl)
				}
			},
		)
	}
}

func TestTemplateCandidates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "single underscore produces prefix candidate",
			input: "ph_abc123",
			expected: []string{
				"ph_*",
			},
		},
		{
			name:  "multiple underscores produces prefix and sandwich candidates",
			input: "ph_phc_abc123",
			expected: []string{
				"ph_*",
				"ph_phc_*",
				"ph_*_abc123",
			},
		},
		{
			name:  "four tokens produces multiple sandwich candidates",
			input: "ph_phc_abc123_posthog",
			expected: []string{
				"ph_*",
				"ph_phc_*",
				"ph_phc_abc123_*",
				"ph_*_abc123_posthog",
				"ph_phc_*_posthog",
			},
		},
		{
			name:  "leading underscore",
			input: "_ga_GB2J3DLBHE",
			expected: []string{
				"_*",
				"_ga_*",
				"_*_GB2J3DLBHE",
			},
		},
		{
			name:  "dash separator",
			input: "c15t-consent-abc123",
			expected: []string{
				"c15t-*",
				"c15t-consent-*",
				"c15t-*-abc123",
			},
		},
		{
			name:     "no separators",
			input:    "PHPSESSID",
			expected: nil,
		},
		{
			name:  "brand with digits",
			input: "auth0_session_abc123",
			expected: []string{
				"auth0_*",
				"auth0_session_*",
				"auth0_*_abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				result := templateCandidates(tt.input)
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}

func TestGlobMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		input   string
		match   bool
	}{
		{
			name:    "prefix glob matches",
			pattern: "ph_phc_*",
			input:   "ph_phc_abc123",
			match:   true,
		},
		{
			name:    "prefix glob does not match different prefix",
			pattern: "ph_phc_*",
			input:   "ph_session_abc123",
			match:   false,
		},
		{
			name:    "suffix glob matches",
			pattern: "*_posthog",
			input:   "ph_phc_abc123_posthog",
			match:   true,
		},
		{
			name:    "suffix glob does not match different suffix",
			pattern: "*_posthog",
			input:   "ph_phc_abc123_analytics",
			match:   false,
		},
		{
			name:    "sandwich glob matches",
			pattern: "ph_phc_*_posthog",
			input:   "ph_phc_abc123_posthog",
			match:   true,
		},
		{
			name:    "sandwich glob does not match wrong prefix",
			pattern: "ph_phc_*_posthog",
			input:   "xx_phc_abc123_posthog",
			match:   false,
		},
		{
			name:    "sandwich glob does not match wrong suffix",
			pattern: "ph_phc_*_posthog",
			input:   "ph_phc_abc123_other",
			match:   false,
		},
		{
			name:    "sandwich glob matches minimal middle",
			pattern: "ph_phc_*_posthog",
			input:   "ph_phc_x_posthog",
			match:   true,
		},
		{
			name:    "sandwich glob matches empty middle",
			pattern: "ph_*_posthog",
			input:   "ph__posthog",
			match:   true,
		},
		{
			name:    "exact match without wildcard",
			pattern: "probo_consent",
			input:   "probo_consent",
			match:   true,
		},
		{
			name:    "no match without wildcard",
			pattern: "probo_consent",
			input:   "probo_consent2",
			match:   false,
		},
		{
			name:    "input shorter than pattern fixed chars does not match",
			pattern: "ph_phc_*_posthog",
			input:   "ph_phc_posthog",
			match:   false,
		},
		{
			name:    "multi-star matches",
			pattern: "ph_*_something_*_end",
			input:   "ph_hash1_something_hash2_end",
			match:   true,
		},
		{
			name:    "multi-star wrong middle segment",
			pattern: "ph_*_something_*_end",
			input:   "ph_hash1_other_hash2_end",
			match:   false,
		},
		{
			name:    "multi-star wrong suffix",
			pattern: "ph_*_something_*_end",
			input:   "ph_hash1_something_hash2_nope",
			match:   false,
		},
		{
			name:    "multi-star with underscore-heavy middle",
			pattern: "a_*_b_*_c",
			input:   "a_x_y_z_b_q_r_c",
			match:   true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.match, globMatch(tt.pattern, tt.input))
			},
		)
	}
}

func TestSplitTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		tokens []string
		seps   []byte
	}{
		{
			name:   "underscore separator",
			input:  "ph_phc_abc",
			tokens: []string{"ph", "phc", "abc"},
			seps:   []byte{'_', '_'},
		},
		{
			name:   "dash separator",
			input:  "c15t-consent-abc",
			tokens: []string{"c15t", "consent", "abc"},
			seps:   []byte{'-', '-'},
		},
		{
			name:   "no separator",
			input:  "PHPSESSID",
			tokens: []string{"PHPSESSID"},
			seps:   nil,
		},
		{
			name:   "mixed separators split both",
			input:  "foo_bar-baz",
			tokens: []string{"foo", "bar", "baz"},
			seps:   []byte{'_', '-'},
		},
		{
			name:   "leading underscores with dash",
			input:  "__Secure-1PSID",
			tokens: []string{"", "", "Secure", "1PSID"},
			seps:   []byte{'_', '_', '-'},
		},
		{
			name:   "UUID preserved as single token",
			input:  "session_550e8400-e29b-41d4-a716-446655440000_data",
			tokens: []string{"session", "550e8400-e29b-41d4-a716-446655440000", "data"},
			seps:   []byte{'_', '_'},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				tokens, seps := splitTokens(tt.input)
				assert.Equal(t, tt.tokens, tokens)
				assert.Equal(t, tt.seps, seps)
			},
		)
	}
}

func TestFindMergeGroups(t *testing.T) {
	t.Parallel()

	oneYear := 365 * 24 * 3600

	makePattern := func(name string, maxAge *int) *coredata.TrackerPattern {
		return &coredata.TrackerPattern{
			Pattern:       name,
			TrackerType:   coredata.TrackerTypeCookie,
			MatchType:     coredata.TrackerPatternMatchTypeExact,
			MaxAgeSeconds: maxAge,
		}
	}

	t.Run(
		"multi-separator names pick longest prefix template",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_abc123", &oneYear),
				makePattern("ph_phc_def456", &oneYear),
				makePattern("ph_phc_ghi789", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"leading separator cookies",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("_ga_ABC123", &oneYear),
				makePattern("_ga_DEF456", &oneYear),
				makePattern("_ga_GHI789", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "_ga_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"brand name with digits in prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("auth0_session_abc123", &oneYear),
				makePattern("auth0_session_def456", &oneYear),
				makePattern("auth0_session_ghi789", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "auth0_session_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"no merge below threshold",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("deadbeef_setting", &oneYear),
				makePattern("something_else", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			assert.Empty(t, groups)
		},
	)

	t.Run(
		"nested prefix resolution prefers longest",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("foo_bar_aaa", &oneYear),
				makePattern("foo_bar_bbb", &oneYear),
				makePattern("foo_bar_ccc", &oneYear),
				makePattern("foo_baz_xxx", &oneYear),
				makePattern("foo_baz_yyy", &oneYear),
				makePattern("foo_baz_zzz", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 2)

			barGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "foo_bar_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, barGroup, 3)

			bazGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "foo_baz_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, bazGroup, 3)
		},
	)

	t.Run(
		"specific prefix wins over broad prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_abc123", &oneYear),
				makePattern("ph_phc_def456", &oneYear),
				makePattern("ph_phc_ghi789", &oneYear),
				makePattern("ph_session_xyz", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"leftover patterns form group under shorter prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_abc123", &oneYear),
				makePattern("ph_phc_def456", &oneYear),
				makePattern("ph_phc_ghi789", &oneYear),
				makePattern("ph_session_aaa", &oneYear),
				makePattern("ph_session_bbb", &oneYear),
				makePattern("ph_session_ccc", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 2)

			phcGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, phcGroup, 3)

			sessionGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_session_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, sessionGroup, 3)
		},
	)

	t.Run(
		"no separators means no merge",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("PHPSESSID", nil),
				makePattern("JSESSIONID", nil),
				makePattern("ASPSESSIONID", nil),
			}

			groups := findMergeGroups(patterns, 3)
			assert.Empty(t, groups)
		},
	)

	t.Run(
		"session and persistent cookies do not merge",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("_ga_ABC123", nil),
				makePattern("_ga_DEF456", nil),
				makePattern("_ga_GHI789", nil),
				makePattern("_ga_JKL012", &oneYear),
				makePattern("_ga_MNO345", &oneYear),
				makePattern("_ga_PQR678", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 2)

			sessionGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "_ga_*", durationBucket: -1}]
			require.True(t, ok)
			assert.Len(t, sessionGroup, 3)

			persistentGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "_ga_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, persistentGroup, 3)
		},
	)

	t.Run(
		"close durations snap to same bucket and merge",
		func(t *testing.T) {
			t.Parallel()

			exactYear := 365 * 24 * 3600
			almostYear := 364 * 24 * 3600

			patterns := coredata.TrackerPatterns{
				makePattern("_ga_ABC123", &exactYear),
				makePattern("_ga_DEF456", &almostYear),
				makePattern("_ga_GHI789", &exactYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)
		},
	)

	t.Run(
		"different durations do not merge",
		func(t *testing.T) {
			t.Parallel()

			oneDay := 24 * 3600
			thirtyDays := 30 * 24 * 3600

			patterns := coredata.TrackerPatterns{
				makePattern("_ga_ABC123", &oneDay),
				makePattern("_ga_DEF456", &oneDay),
				makePattern("_ga_GHI789", &oneDay),
				makePattern("_ga_JKL012", &thirtyDays),
				makePattern("_ga_MNO345", &thirtyDays),
				makePattern("_ga_PQR678", &thirtyDays),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 2)

			dayGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "_ga_*", durationBucket: durationBucket(&oneDay)}]
			require.True(t, ok)
			assert.Len(t, dayGroup, 3)

			monthGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "_ga_*", durationBucket: durationBucket(&thirtyDays)}]
			require.True(t, ok)
			assert.Len(t, monthGroup, 3)
		},
	)

	t.Run(
		"sandwich pattern discovered from shared prefix and suffix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_abc123_posthog", &oneYear),
				makePattern("ph_phc_def456_posthog", &oneYear),
				makePattern("ph_phc_ghi789_posthog", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*_posthog", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"sandwich pattern wins over shorter prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_abc123_posthog", &oneYear),
				makePattern("ph_phc_def456_posthog", &oneYear),
				makePattern("ph_phc_ghi789_posthog", &oneYear),
				makePattern("ph_phc_other", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*_posthog", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"single pattern with hash normalized via heuristic",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_window_id", nil),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*_window_id", durationBucket: -1}]
			require.True(t, ok)
			assert.Len(t, group, 1)
		},
	)

	t.Run(
		"same hash different suffixes produce separate heuristic globs",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_window_id", nil),
				makePattern("ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_posthog", nil),
				makePattern("ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_primary_window_exists", nil),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 3)

			windowGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*_window_id", durationBucket: -1}]
			require.True(t, ok)
			assert.Len(t, windowGroup, 1)

			posthogGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*_posthog", durationBucket: -1}]
			require.True(t, ok)
			assert.Len(t, posthogGroup, 1)

			primaryGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*_primary_window_exists", durationBucket: -1}]
			require.True(t, ok)
			assert.Len(t, primaryGroup, 1)
		},
	)

	t.Run(
		"patterns without variable tokens still require statistical threshold",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("foo_bar_aaa", &oneYear),
				makePattern("foo_bar_bbb", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			assert.Empty(t, groups)
		},
	)

	t.Run(
		"heuristic and statistical patterns coexist",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_window_id", &oneYear),
				makePattern("foo_bar_aaa", &oneYear),
				makePattern("foo_bar_bbb", &oneYear),
				makePattern("foo_bar_ccc", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 2)

			heuristicGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "ph_phc_*_window_id", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, heuristicGroup, 1)

			statGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, template: "foo_bar_*", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, statGroup, 3)
		},
	)

	t.Run(
		"secure prefix cookies do not produce heuristic glob",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("__Secure-1PSID", &oneYear),
				makePattern("__Secure-1PSIDTS", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			assert.Empty(t, groups)
		},
	)
}

func TestDurationBucket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		maxAge   *int
		expected int
	}{
		{
			name:     "nil is session",
			maxAge:   nil,
			expected: -1,
		},
		{
			name:     "zero is session",
			maxAge:   new(0),
			expected: -1,
		},
		{
			name:     "negative is session",
			maxAge:   new(-1),
			expected: -1,
		},
		{
			name:     "exact 1 year",
			maxAge:   new(365 * 24 * 3600),
			expected: 365 * 24 * 3600,
		},
		{
			name:     "364 days snaps to 1 year",
			maxAge:   new(364 * 24 * 3600),
			expected: 365 * 24 * 3600,
		},
		{
			name:     "exact 30 days",
			maxAge:   new(30 * 24 * 3600),
			expected: 30 * 24 * 3600,
		},
		{
			name:     "exact 1 day",
			maxAge:   new(24 * 3600),
			expected: 24 * 3600,
		},
		{
			name:     "23h snaps to 1 day",
			maxAge:   new(23 * 3600),
			expected: 24 * 3600,
		},
		{
			name:     "exact 1 hour",
			maxAge:   new(3600),
			expected: 3600,
		},
		{
			name:     "58 minutes snaps to 1 hour",
			maxAge:   new(58 * 60),
			expected: 3600,
		},
		{
			name:     "exact 5 minutes",
			maxAge:   new(5 * 60),
			expected: 5 * 60,
		},
		{
			name:     "1 day and 30 days are different buckets",
			maxAge:   new(24 * 3600),
			expected: 24 * 3600,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				result := durationBucket(tt.maxAge)
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}
