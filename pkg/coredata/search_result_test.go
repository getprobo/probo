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

package coredata

import (
	"maps"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestSearchResultsSQL_NamedArgsMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		query     string
		selectors []TagSelector
	}{
		{name: "title", query: "UniqueTitle"},
		{name: "tag", selectors: []TagSelector{{Name: "SearchProduction"}}},
		{
			name:  "title and tags",
			query: "UniqueTitle",
			selectors: []TagSelector{
				{Name: "SearchProduction"},
				{Name: "SearchCompliance"},
			},
		},
		{
			name:      "punctuation and tag",
			query:     "---",
			selectors: []TagSelector{{Name: "SearchProduction"}},
		},
	}

	scope := NewScope(gid.NewTenantID(), gid.Nil)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			order := page.OrderBy[SearchResultOrderField]{
				Field:     SearchResultOrderFieldID,
				Direction: page.OrderDirectionAsc,
			}
			if hasSearchTerms(tc.query) {
				order = page.OrderBy[SearchResultOrderField]{
					Field:     SearchResultOrderFieldRank,
					Direction: page.OrderDirectionDesc,
				}
			}

			cursor := page.NewCursor(page.DefaultCursorSize, nil, page.Head, order)

			selectArgs := searchResultArgs(
				gid.Nil,
				tc.selectors,
				[]uint16{DocumentEntityType},
				tc.query,
			)
			maps.Copy(selectArgs, scope.SQLArguments())
			maps.Copy(selectArgs, cursor.SQLArguments())

			selectSQL := searchResultsSelectSQL(
				tc.query,
				scope.SQLFragment(),
				cursor.SQLFragment(),
			)
			_, _, err := pgx.StrictNamedArgs(selectArgs).RewriteQuery(
				t.Context(),
				nil,
				selectSQL,
				nil,
			)
			require.NoError(t, err)

			countArgs := searchResultArgs(
				gid.Nil,
				tc.selectors,
				[]uint16{DocumentEntityType},
				tc.query,
			)
			maps.Copy(countArgs, scope.SQLArguments())

			countSQL := searchResultsCountSQL(tc.query, scope.SQLFragment())
			_, _, err = pgx.StrictNamedArgs(countArgs).RewriteQuery(
				t.Context(),
				nil,
				countSQL,
				nil,
			)
			require.NoError(t, err)
		})
	}
}
