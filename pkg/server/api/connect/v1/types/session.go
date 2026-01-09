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
	SessionOrderBy OrderBy[coredata.SessionOrderField]

	SessionConnection struct {
		TotalCount int
		Edges      []*SessionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewSessionConnection(
	p *page.Page[*coredata.Session, coredata.SessionOrderField],
	resolver any,
	parentID gid.GID,
) *SessionConnection {
	edges := make([]*SessionEdge, len(p.Data))
	for i, session := range p.Data {
		edges[i] = NewSessionEdge(session, p.Cursor.OrderBy.Field)
	}

	return &SessionConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewSessionEdge(session *coredata.Session, orderField coredata.SessionOrderField) *SessionEdge {
	return &SessionEdge{
		Node:   NewSession(session),
		Cursor: session.CursorKey(orderField),
	}
}

func NewSession(session *coredata.Session) *Session {
	return &Session{
		ID: session.ID,
		Identity: &Identity{
			ID: session.IdentityID,
		},
		IPAddress: session.IPAddress.String(),
		UserAgent: session.UserAgent,
		UpdatedAt: session.UpdatedAt,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiredAt,
	}
}
