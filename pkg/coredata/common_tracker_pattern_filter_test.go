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

package coredata_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// TestCommonTrackerPatternFilter_PatternKeywordIgnoresDescription pins why
// --pattern-keyword exists.
//
// Keyword matches the description too, and the bulk actions that select rows
// this way also blank descriptions, so the same selection returns fewer rows
// each time it runs. An operator comparing a list against a mark sees two
// different populations and cannot tell which rows were acted on. Matching the
// pattern key, which nothing rewrites, keeps the selection stable.
func TestCommonTrackerPatternFilter_PatternKeywordIgnoresDescription(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	token := gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// One row carries the needle in its key, the other only in its
	// description.
	inKey := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     "needle_" + token,
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Description: "unrelated prose",
		Confidence:  1,
		Attribution: coredata.CommonTrackerPatternAttributionUndetermined,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	inDescription := inKey
	inDescription.ID = gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType)
	inDescription.Pattern = "plain_" + token
	inDescription.Description = "set by needle_" + token

	insertCommonTrackerPattern(t, ctx, client, inKey)
	insertCommonTrackerPattern(t, ctx, client, inDescription)

	load := func(apply func(*coredata.CommonTrackerPatternFilter)) []string {
		t.Helper()

		filter := coredata.NewCommonTrackerPatternFilter()
		apply(filter)

		var patterns []string

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, c pg.Querier) error {
			rows, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.CommonTrackerPatternOrderField]{
					Field:     coredata.CommonTrackerPatternOrderFieldPattern,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.CommonTrackerPatternOrderField]) ([]*coredata.CommonTrackerPattern, error) {
					var batch coredata.CommonTrackerPatterns
					if err := batch.Load(ctx, c, cursor, filter); err != nil {
						return nil, err
					}

					return batch, nil
				},
			)
			if err != nil {
				return err
			}

			for _, r := range rows {
				// The catalog is global, so restrict to this test's rows.
				if strings.Contains(r.Pattern, token) {
					patterns = append(patterns, r.Pattern)
				}
			}

			return nil
		}))

		return patterns
	}

	needle := "needle_" + token

	assert.ElementsMatch(
		t,
		[]string{inKey.Pattern, inDescription.Pattern},
		load(func(f *coredata.CommonTrackerPatternFilter) { f.WithKeyword(&needle) }),
		"keyword matches the description too",
	)

	assert.Equal(
		t,
		[]string{inKey.Pattern},
		load(func(f *coredata.CommonTrackerPatternFilter) { f.WithPatternKeyword(&needle) }),
		"pattern keyword must ignore the description",
	)

	// Blank both descriptions, as every terminal-verdict action does.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var ps coredata.CommonTrackerPatterns

		_, err := ps.ClearDescriptionByIDs(ctx, tx, []gid.GID{inKey.ID, inDescription.ID})

		return err
	}))

	assert.Equal(
		t,
		[]string{inKey.Pattern},
		load(func(f *coredata.CommonTrackerPatternFilter) { f.WithKeyword(&needle) }),
		"keyword drifts once the descriptions it matched are cleared",
	)

	assert.Equal(
		t,
		[]string{inKey.Pattern},
		load(func(f *coredata.CommonTrackerPatternFilter) { f.WithPatternKeyword(&needle) }),
		"pattern keyword must return the same rows before and after",
	)
}

// TestCommonTrackerPatternFilter_PatternKeywordTreatsWildcardsLiterally
// pins that %, _, and backslash are the literal substring the flag documents,
// not ILIKE wildcards. Unescaped, "%" would match every row — which on a
// terminal bulk action would mark the entire catalog.
func TestCommonTrackerPatternFilter_PatternKeywordTreatsWildcardsLiterally(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	// GIDs are base64url and can contain '_'. Strip it so the unique suffix
	// cannot make every row match a literal underscore search.
	token := strings.ReplaceAll(
		gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String(),
		"_",
		"X",
	)
	now := time.Now().UTC().Truncate(time.Microsecond)

	insert := func(pattern string) string {
		t.Helper()

		row := coredata.CommonTrackerPattern{
			ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
			TrackerType: coredata.TrackerTypeCookie,
			Pattern:     pattern,
			MatchType:   coredata.TrackerPatternMatchTypeExact,
			Confidence:  1,
			Attribution: coredata.CommonTrackerPatternAttributionUndetermined,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		insertCommonTrackerPattern(t, ctx, client, row)

		return pattern
	}

	// The decoy has none of the metacharacters. Unescaped, "%" matches
	// every row and "_" matches every non-empty row; those regressions
	// would include this pattern.
	decoy := insert("plain" + token)
	hasPercent := insert("has%pct" + token)
	hasUnderscore := insert("has_und" + token)
	hasBackslash := insert(`has\bs` + token)

	load := func(keyword string) []string {
		t.Helper()

		filter := coredata.NewCommonTrackerPatternFilter()
		filter.WithPatternKeyword(&keyword)

		var patterns []string

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, c pg.Querier) error {
			rows, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.CommonTrackerPatternOrderField]{
					Field:     coredata.CommonTrackerPatternOrderFieldPattern,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.CommonTrackerPatternOrderField]) ([]*coredata.CommonTrackerPattern, error) {
					var batch coredata.CommonTrackerPatterns
					if err := batch.Load(ctx, c, cursor, filter); err != nil {
						return nil, err
					}

					return batch, nil
				},
			)
			if err != nil {
				return err
			}

			for _, r := range rows {
				if strings.Contains(r.Pattern, token) {
					patterns = append(patterns, r.Pattern)
				}
			}

			return nil
		}))

		return patterns
	}

	assert.Equal(t, []string{hasPercent}, load("%"), "a bare % must not match the whole catalog")
	assert.Equal(t, []string{hasUnderscore}, load("_"), "a bare _ must match only a literal underscore")
	assert.Equal(t, []string{hasBackslash}, load(`\`), "a backslash must match as a literal")
	assert.Equal(t, []string{hasPercent}, load("has%pct"), "a % inside a needle must stay literal")
	assert.NotContains(t, load("%"), decoy)
	assert.NotContains(t, load("_"), decoy)
	assert.NotContains(t, load(`\`), decoy)
}
