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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewSnapshot(s *coredata.Snapshot) *Snapshot {
	return &Snapshot{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Name:           s.Name,
		Type:           s.Type,
		Description:    s.Description,
		CreatedAt:      s.CreatedAt,
	}
}

func NewListSnapshotsOutput(snapshotPage *page.Page[*coredata.Snapshot, coredata.SnapshotOrderField]) ListSnapshotsOutput {
	snapshots := make([]*Snapshot, 0, len(snapshotPage.Data))
	for _, s := range snapshotPage.Data {
		snapshots = append(snapshots, NewSnapshot(s))
	}

	var nextCursor *page.CursorKey
	if len(snapshotPage.Data) > 0 {
		cursorKey := snapshotPage.Data[len(snapshotPage.Data)-1].CursorKey(snapshotPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListSnapshotsOutput{
		NextCursor: nextCursor,
		Snapshots:  snapshots,
	}
}
