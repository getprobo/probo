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
