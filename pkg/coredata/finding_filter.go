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
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type (
	FindingFilter struct {
		kind     *FindingKind
		status   *FindingStatus
		priority *FindingPriority
		ownerID  *gid.GID
	}
)

func NewFindingFilter(
	kind *FindingKind,
	status *FindingStatus,
	priority *FindingPriority,
	ownerID *gid.GID,
) *FindingFilter {
	return &FindingFilter{
		kind:     kind,
		status:   status,
		priority: priority,
		ownerID:  ownerID,
	}
}

func (f *FindingFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"has_kind_filter":     false,
		"filter_kind":         nil,
		"has_status_filter":   false,
		"filter_status":       nil,
		"has_priority_filter": false,
		"filter_priority":     nil,
		"has_owner_filter":    false,
		"filter_owner_id":     nil,
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
