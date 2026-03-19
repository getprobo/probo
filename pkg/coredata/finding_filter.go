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

package coredata

import (
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type (
	FindingFilter struct {
		snapshotID **gid.GID
		kind       *FindingKind
		status     *FindingStatus
		priority   *FindingPriority
		ownerID    *gid.GID
	}
)

func NewFindingFilter(
	snapshotID **gid.GID,
	kind *FindingKind,
	status *FindingStatus,
	priority *FindingPriority,
	ownerID *gid.GID,
) *FindingFilter {
	return &FindingFilter{
		snapshotID: snapshotID,
		kind:       kind,
		status:     status,
		priority:   priority,
		ownerID:    ownerID,
	}
}

func (f *FindingFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"has_snapshot_filter": false,
		"filter_snapshot_id":  nil,
		"has_kind_filter":     false,
		"filter_kind":         nil,
		"has_status_filter":   false,
		"filter_status":       nil,
		"has_priority_filter": false,
		"filter_priority":     nil,
		"has_owner_filter":    false,
		"filter_owner_id":     nil,
	}

	if f.snapshotID != nil {
		args["has_snapshot_filter"] = true
		if *f.snapshotID != nil {
			args["filter_snapshot_id"] = **f.snapshotID
		}
	}

	if f.kind != nil {
		args["has_kind_filter"] = true
		args["filter_kind"] = string(*f.kind)
	}

	if f.status != nil {
		args["has_status_filter"] = true
		args["filter_status"] = string(*f.status)
	}

	if f.priority != nil {
		args["has_priority_filter"] = true
		args["filter_priority"] = string(*f.priority)
	}

	if f.ownerID != nil {
		args["has_owner_filter"] = true
		args["filter_owner_id"] = *f.ownerID
	}

	return args
}

func (f *FindingFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @has_snapshot_filter::boolean = false THEN TRUE
        WHEN @has_snapshot_filter::boolean = true AND @filter_snapshot_id::text IS NOT NULL THEN
            snapshot_id = @filter_snapshot_id::text
        WHEN @has_snapshot_filter::boolean = true AND @filter_snapshot_id::text IS NULL THEN
            snapshot_id IS NULL
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_kind_filter::boolean = false THEN TRUE
        WHEN @has_kind_filter::boolean = true THEN
            kind = @filter_kind::findings_kind
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_status_filter::boolean = false THEN TRUE
        WHEN @has_status_filter::boolean = true THEN
            status = @filter_status::findings_status
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_priority_filter::boolean = false THEN TRUE
        WHEN @has_priority_filter::boolean = true THEN
            priority = @filter_priority::findings_priority
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_owner_filter::boolean = false THEN TRUE
        WHEN @has_owner_filter::boolean = true THEN
            owner_id = @filter_owner_id::text
        ELSE TRUE
    END
)`
}
