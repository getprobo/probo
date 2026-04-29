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
)

func TestSeparatorPrefixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single underscore",
			input:    "ph_abc123",
			expected: []string{"ph_"},
		},
		{
			name:     "multiple underscores",
			input:    "ph_phc_abc123def456",
			expected: []string{"ph_", "ph_phc_"},
		},
		{
			name:     "leading underscore",
			input:    "_ga_GB2J3DLBHE",
			expected: []string{"_", "_ga_"},
		},
		{
			name:     "dash separator",
			input:    "c15t-consent-abc123",
			expected: []string{"c15t-", "c15t-consent-"},
		},
		{
			name:     "mixed separators",
			input:    "_gat_UA-12345-1",
			expected: []string{"_", "_gat_", "_gat_UA-", "_gat_UA-12345-"},
		},
		{
			name:     "no separators",
			input:    "PHPSESSID",
			expected: nil,
		},
		{
			name:     "brand with digits",
			input:    "auth0_session_abc123",
			expected: []string{"auth0_", "auth0_session_"},
		},
		{
			name:     "woocommerce session",
			input:    "wp_woocommerce_session_f919208d949256bc",
			expected: []string{"wp_", "wp_woocommerce_", "wp_woocommerce_session_"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				result := separatorPrefixes(tt.input)
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}

func TestFindMergeGroups(t *testing.T) {
	t.Parallel()

	makePattern := func(name string) *coredata.CookiePattern {
		return &coredata.CookiePattern{
			Pattern:   name,
			MatchType: coredata.CookiePatternMatchTypeExact,
		}
	}

	t.Run(
		"multi-separator names pick longest prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("ph_phc_abc123"),
				makePattern("ph_phc_def456"),
				makePattern("ph_phc_ghi789"),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups["ph_phc_"]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"leading separator cookies",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("_ga_ABC123"),
				makePattern("_ga_DEF456"),
				makePattern("_ga_GHI789"),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups["_ga_"]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"brand name with digits in prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("auth0_session_abc123"),
				makePattern("auth0_session_def456"),
				makePattern("auth0_session_ghi789"),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups["auth0_session_"]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"no merge below threshold",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("deadbeef_setting"),
				makePattern("something_else"),
			}

			groups := findMergeGroups(patterns, 3)
			assert.Empty(t, groups)
		},
	)

	t.Run(
		"nested prefix resolution prefers longest",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("foo_bar_aaa"),
				makePattern("foo_bar_bbb"),
				makePattern("foo_bar_ccc"),
				makePattern("foo_baz_xxx"),
				makePattern("foo_baz_yyy"),
				makePattern("foo_baz_zzz"),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 2)

			barGroup, ok := groups["foo_bar_"]
			require.True(t, ok)
			assert.Len(t, barGroup, 3)

			bazGroup, ok := groups["foo_baz_"]
			require.True(t, ok)
			assert.Len(t, bazGroup, 3)
		},
	)

	t.Run(
		"specific prefix wins over broad prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("ph_phc_abc123"),
				makePattern("ph_phc_def456"),
				makePattern("ph_phc_ghi789"),
				makePattern("ph_session_xyz"),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups["ph_phc_"]
			require.True(t, ok)
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"skips non-exact patterns",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("ph_phc_abc123"),
				makePattern("ph_phc_def456"),
				makePattern("ph_phc_ghi789"),
				{
					Pattern:   "ph_phc_",
					MatchType: coredata.CookiePatternMatchTypePrefix,
				},
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group := groups["ph_phc_"]
			assert.Len(t, group, 3)
		},
	)

	t.Run(
		"leftover patterns form group under shorter prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("ph_phc_abc123"),
				makePattern("ph_phc_def456"),
				makePattern("ph_phc_ghi789"),
				makePattern("ph_session_aaa"),
				makePattern("ph_session_bbb"),
				makePattern("ph_session_ccc"),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 2)

			phcGroup, ok := groups["ph_phc_"]
			require.True(t, ok)
			assert.Len(t, phcGroup, 3)

			sessionGroup, ok := groups["ph_session_"]
			require.True(t, ok)
			assert.Len(t, sessionGroup, 3)
		},
	)

	t.Run(
		"no separators means no merge",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.CookiePatterns{
				makePattern("PHPSESSID"),
				makePattern("JSESSIONID"),
				makePattern("ASPSESSIONID"),
			}

			groups := findMergeGroups(patterns, 3)
			assert.Empty(t, groups)
		},
	)
}
