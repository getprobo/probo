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
	"context"
	"fmt"
	"maps"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	SearchResult struct {
		ID    gid.GID `db:"id"`
		Title string  `db:"title"`
		Rank  float64 `db:"rank"`
	}

	SearchResults []*SearchResult
)

func (s SearchResult) CursorKey(orderBy SearchResultOrderField) page.CursorKey {
	switch orderBy {
	case SearchResultOrderFieldID:
		return page.NewCursorKey(s.ID, s.ID.String())
	case SearchResultOrderFieldRank:
		return page.NewCursorKey(s.ID, s.Rank)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// hasSearchTerms reports whether query contains a letter or digit. A query
// made only of separators tokenizes to nothing and matches every row.
func hasSearchTerms(query string) bool {
	return strings.IndexFunc(
		query,
		func(r rune) bool {
			return unicode.IsLetter(r) || unicode.IsNumber(r)
		},
	) >= 0
}

func searchResourceFilterSQL() string {
	return `
(
	CASE
		WHEN cardinality(@entity_types::integer[]) = 0 THEN TRUE
		ELSE entity_type = ANY(@entity_types::integer[])
	END
	AND
	CASE
		WHEN @selector_count::integer = 0 THEN TRUE
		ELSE resource_id IN (
			SELECT
				resource_id
			FROM
				tag_assignments
			WHERE
				organization_id = @organization_id
				AND tag_id IN (
					SELECT
						id
					FROM
						tags
					WHERE
						organization_id = @organization_id
						AND name = ANY(@selector_names::text[])
				)
			GROUP BY
				resource_id
			HAVING
				COUNT(DISTINCT tag_id) = @selector_count
		)
	END
)`
}

func searchResultArgs(
	organizationID gid.GID,
	selectors []TagSelector,
	entityTypes []uint16,
	query string,
) pgx.StrictNamedArgs {
	names := make([]string, len(selectors))
	for i, selector := range selectors {
		names[i] = selector.Name
	}

	typed := make([]int32, len(entityTypes))
	for i, entityType := range entityTypes {
		typed[i] = int32(entityType)
	}

	// Tag-only SQL omits @query. StrictNamedArgs rejects an argument the
	// query text does not reference.
	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"selector_names":  names,
		"selector_count":  len(selectors),
		"entity_types":    typed,
	}
	if hasSearchTerms(query) {
		args["query"] = query
	}

	return args
}

// The text match stays a bare search_vector @@ predicate. Wrapping it in
// CASE or OR keeps it out of the generic plan's index condition.
func searchResultsSelectSQL(query string, scopeSQL string, cursorSQL string) string {
	var q string
	if hasSearchTerms(query) {
		q = `
SELECT
	id,
	title,
	rank
FROM (
	SELECT
		resource_id AS id,
		title,
		COALESCE(ts_rank(
			search_vector,
			(
				SELECT to_tsquery(
					'simple',
					string_agg(quote_literal(lexeme) || ':*', ' & ')
				)
				FROM unnest(
					to_tsvector(
						'simple',
						translate(@query::text, '-+', '  ')
					)
				) AS lex(lexeme, positions)
			)
		), 0)::float8 AS rank
	FROM
		search_resources
	WHERE
		%s
		AND organization_id = @organization_id
		AND %s
		AND search_vector @@ (
			SELECT to_tsquery(
				'simple',
				string_agg(quote_literal(lexeme) || ':*', ' & ')
			)
			FROM unnest(
				to_tsvector(
					'simple',
					translate(@query::text, '-+', '  ')
				)
			) AS lex(lexeme, positions)
		)
) matched
WHERE
	%s
`
	} else {
		q = `
SELECT
	id,
	title,
	rank
FROM (
	SELECT
		resource_id AS id,
		title,
		0::float8 AS rank
	FROM
		search_resources
	WHERE
		%s
		AND organization_id = @organization_id
		AND %s
) matched
WHERE
	%s
`
	}

	return fmt.Sprintf(q, scopeSQL, searchResourceFilterSQL(), cursorSQL)
}

func searchResultsCountSQL(query string, scopeSQL string) string {
	var q string
	if hasSearchTerms(query) {
		q = `
SELECT
	COUNT(*)
FROM
	search_resources
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
	AND search_vector @@ (
		SELECT to_tsquery(
			'simple',
			string_agg(quote_literal(lexeme) || ':*', ' & ')
		)
		FROM unnest(
			to_tsvector(
				'simple',
				translate(@query::text, '-+', '  ')
			)
		) AS lex(lexeme, positions)
	)
`
	} else {
		q = `
SELECT
	COUNT(*)
FROM
	search_resources
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`
	}

	return fmt.Sprintf(q, scopeSQL, searchResourceFilterSQL())
}

func (rs *SearchResults) LoadByOrganizationIDAndSelectors(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	selectors []TagSelector,
	entityTypes []uint16,
	query string,
	cursor *page.Cursor[SearchResultOrderField],
) error {
	if len(selectors) == 0 && query == "" {
		*rs = nil

		return nil
	}

	args := searchResultArgs(organizationID, selectors, entityTypes, query)
	maps.Copy(args, scope.SQLArguments())

	q := searchResultsSelectSQL(query, scope.SQLFragment(), cursor.SQLFragment())

	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query search results: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SearchResult])
	if err != nil {
		return fmt.Errorf("cannot collect search results: %w", err)
	}

	*rs = results

	return nil
}

func (rs *SearchResults) CountByOrganizationIDAndSelectors(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	selectors []TagSelector,
	entityTypes []uint16,
	query string,
) (int, error) {
	if len(selectors) == 0 && query == "" {
		return 0, nil
	}

	q := searchResultsCountSQL(query, scope.SQLFragment())

	args := searchResultArgs(organizationID, selectors, entityTypes, query)
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count search results: %w", err)
	}

	return count, nil
}
