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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	MeetingOrderBy OrderBy[coredata.MeetingOrderField]

	MeetingConnection struct {
		TotalCount int
		Edges      []*MeetingEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewMeetingConnection(
	p *page.Page[*coredata.Meeting, coredata.MeetingOrderField],
	parentType any,
	parentID gid.GID,
) *MeetingConnection {
	var edges = make([]*MeetingEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewMeetingEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &MeetingConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewMeetingEdges(meetings []*coredata.Meeting, orderBy coredata.MeetingOrderField) []*MeetingEdge {
	edges := make([]*MeetingEdge, len(meetings))

	for i := range edges {
		edges[i] = NewMeetingEdge(meetings[i], orderBy)
	}

	return edges
}

func NewMeetingEdge(meeting *coredata.Meeting, orderBy coredata.MeetingOrderField) *MeetingEdge {
	return &MeetingEdge{
		Cursor: meeting.CursorKey(orderBy),
		Node:   NewMeeting(meeting),
	}
}

func NewMeeting(meeting *coredata.Meeting) *Meeting {
	return &Meeting{
		ID:   meeting.ID,
		Name: meeting.Name,
		Organization: &Organization{
			ID: meeting.OrganizationID,
		},
		Date:      meeting.Date,
		Minutes:   meeting.Minutes,
		CreatedAt: meeting.CreatedAt,
		UpdatedAt: meeting.UpdatedAt,
	}
}
