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
		"multi-separator names pick longest prefix",
		func(t *testing.T) {
			t.Parallel()

			patterns := coredata.TrackerPatterns{
				makePattern("ph_phc_abc123", &oneYear),
				makePattern("ph_phc_def456", &oneYear),
				makePattern("ph_phc_ghi789", &oneYear),
			}

			groups := findMergeGroups(patterns, 3)
			require.Len(t, groups, 1)

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "ph_phc_", durationBucket: durationBucket(&oneYear)}]
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

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "_ga_", durationBucket: durationBucket(&oneYear)}]
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

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "auth0_session_", durationBucket: durationBucket(&oneYear)}]
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

			barGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "foo_bar_", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, barGroup, 3)

			bazGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "foo_baz_", durationBucket: durationBucket(&oneYear)}]
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

			group, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "ph_phc_", durationBucket: durationBucket(&oneYear)}]
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

			phcGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "ph_phc_", durationBucket: durationBucket(&oneYear)}]
			require.True(t, ok)
			assert.Len(t, phcGroup, 3)

			sessionGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "ph_session_", durationBucket: durationBucket(&oneYear)}]
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

			sessionGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "_ga_", durationBucket: -1}]
			require.True(t, ok)
			assert.Len(t, sessionGroup, 3)

			persistentGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "_ga_", durationBucket: durationBucket(&oneYear)}]
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

			dayGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "_ga_", durationBucket: durationBucket(&oneDay)}]
			require.True(t, ok)
			assert.Len(t, dayGroup, 3)

			monthGroup, ok := groups[mergeGroupKey{categoryID: gid.Nil, trackerType: coredata.TrackerTypeCookie, prefix: "_ga_", durationBucket: durationBucket(&thirtyDays)}]
			require.True(t, ok)
			assert.Len(t, monthGroup, 3)
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
