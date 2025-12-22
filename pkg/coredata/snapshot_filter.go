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
	"time"

	"github.com/jackc/pgx/v5"
)

type (
	SnapshotFilter struct {
		snapshotType *SnapshotsType
		beforeDate   *time.Time
	}
)

func NewSnapshotFilter(snapshotType *SnapshotsType) *SnapshotFilter {
	return &SnapshotFilter{
		snapshotType: snapshotType,
	}
}

func (f *SnapshotFilter) WithBeforeDate(beforeDate *time.Time) *SnapshotFilter {
	f.beforeDate = beforeDate
	return f
}

func (f *SnapshotFilter) SQLArguments() pgx.NamedArgs {
	args := pgx.NamedArgs{
		"filter_snapshot_type": f.snapshotType,
		"filter_before_date":   f.beforeDate,
	}

	return args
}

func (f *SnapshotFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_snapshot_type::snapshots_type IS NOT NULL THEN
			type = @filter_snapshot_type::snapshots_type
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_before_date::timestamptz IS NOT NULL THEN
			created_at <= @filter_before_date::timestamptz
		ELSE TRUE
	END
)`
}
