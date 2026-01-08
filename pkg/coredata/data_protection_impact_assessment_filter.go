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
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type (
	DataProtectionImpactAssessmentFilter struct {
		snapshotID **gid.GID
	}
)

func NewDataProtectionImpactAssessmentFilter(snapshotID **gid.GID) *DataProtectionImpactAssessmentFilter {
	return &DataProtectionImpactAssessmentFilter{
		snapshotID: snapshotID,
	}
}

func (f *DataProtectionImpactAssessmentFilter) SQLArguments() pgx.NamedArgs {
	args := pgx.NamedArgs{}

	if f.snapshotID != nil && *f.snapshotID != nil {
		args["filter_snapshot_id"] = **f.snapshotID
	}

	return args
}

func (f *DataProtectionImpactAssessmentFilter) SQLFragment() string {
	if f.snapshotID == nil {
		return "TRUE"
	}

	if *f.snapshotID == nil {
		return "snapshot_id IS NULL"
	} else {
		return "snapshot_id = @filter_snapshot_id"
	}
}

func (f *DataProtectionImpactAssessmentFilter) SnapshotID() *gid.GID {
	if f.snapshotID == nil || *f.snapshotID == nil {
		return nil
	}
	return *f.snapshotID
}
