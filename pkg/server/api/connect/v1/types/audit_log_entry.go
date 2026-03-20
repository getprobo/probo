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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AuditLogEntryOrderBy OrderBy[coredata.AuditLogEntryOrderField]

	AuditLogEntryConnection struct {
		TotalCount int
		Edges      []*AuditLogEntryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.AuditLogEntryFilter
	}
)

func NewAuditLogEntryConnection(
	p *page.Page[*coredata.AuditLogEntry, coredata.AuditLogEntryOrderField],
	resolver any,
	parentID gid.GID,
	filter *coredata.AuditLogEntryFilter,
) *AuditLogEntryConnection {
	edges := make([]*AuditLogEntryEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAuditLogEntryEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AuditLogEntryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewAuditLogEntryEdge(e *coredata.AuditLogEntry, orderBy coredata.AuditLogEntryOrderField) *AuditLogEntryEdge {
	return &AuditLogEntryEdge{
		Cursor: e.CursorKey(orderBy),
		Node:   NewAuditLogEntry(e),
	}
}

func NewAuditLogEntry(e *coredata.AuditLogEntry) *AuditLogEntry {
	var metadata *string
	if len(e.Metadata) > 0 {
		metadata = new(string(e.Metadata))
	}

	return &AuditLogEntry{
		ID: e.ID,
		Organization: &Organization{
			ID: e.OrganizationID,
		},
		ActorID:      e.ActorID,
		ActorType:    e.ActorType,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		Metadata:     metadata,
		CreatedAt:    e.CreatedAt,
	}
}
