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
	PersonalAPIKeyOrderBy OrderBy[coredata.PersonalAPIKeyOrderField]

	PersonalAPIKeyConnection struct {
		TotalCount int
		Edges      []*PersonalAPIKeyEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewPersonalAPIKeyConnection(
	p *page.Page[*coredata.PersonalAPIKey, coredata.PersonalAPIKeyOrderField],
	resolver any,
	parentID gid.GID,
) *PersonalAPIKeyConnection {
	edges := make([]*PersonalAPIKeyEdge, len(p.Data))
	for i, personalAPIKey := range p.Data {
		edges[i] = NewPersonalAPIKeyEdge(personalAPIKey, p.Cursor.OrderBy.Field)
	}

	return &PersonalAPIKeyConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewPersonalAPIKeyEdge(personalAPIKey *coredata.PersonalAPIKey, orderField coredata.PersonalAPIKeyOrderField) *PersonalAPIKeyEdge {
	return &PersonalAPIKeyEdge{
		Node:   NewPersonalAPIKey(personalAPIKey),
		Cursor: personalAPIKey.CursorKey(orderField),
	}
}

func NewPersonalAPIKey(personalAPIKey *coredata.PersonalAPIKey) *PersonalAPIKey {
	return &PersonalAPIKey{
		ID:         personalAPIKey.ID,
		Name:       personalAPIKey.Name,
		ExpiresAt:  personalAPIKey.ExpiresAt,
		LastUsedAt: personalAPIKey.LastUsedAt,
		CreatedAt:  personalAPIKey.CreatedAt,
	}
}
