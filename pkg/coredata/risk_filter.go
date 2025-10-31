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

package coredata

import (
	"go.probo.inc/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
)

type (
	RiskFilter struct {
		query      *string
		snapshotID **gid.GID
	}
)

func NewRiskFilter(query *string, snapshotID **gid.GID) *RiskFilter {
	return &RiskFilter{
		query:      query,
		snapshotID: snapshotID,
	}
}

func (f *RiskFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"query": f.query,
	}

	if f.snapshotID == nil {
		args["has_snapshot_filter"] = false
		args["filter_snapshot_id"] = nil
	} else if *f.snapshotID == nil {
		args["has_snapshot_filter"] = true
		args["filter_snapshot_id"] = nil
	} else {
		args["has_snapshot_filter"] = true
		args["filter_snapshot_id"] = **f.snapshotID
	}

	return args
}

func (f *RiskFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @query::text IS NOT NULL AND @query::text != '' THEN
			search_vector @@ (
				SELECT to_tsquery('simple', string_agg(lexeme || ':*', ' & '))
				FROM unnest(regexp_split_to_array(trim(@query), '\s+')) AS lexeme
			)
		ELSE TRUE
	END
	AND
	CASE
		WHEN @has_snapshot_filter::boolean = false THEN TRUE
		WHEN @has_snapshot_filter::boolean = true AND @filter_snapshot_id::text IS NOT NULL THEN
			snapshot_id = @filter_snapshot_id::text
		WHEN @has_snapshot_filter::boolean = true AND @filter_snapshot_id::text IS NULL THEN
			snapshot_id IS NULL
		ELSE TRUE
	END
)`
}
