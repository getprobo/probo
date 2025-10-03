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
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
)

type (
	EvidenceOrderBy OrderBy[coredata.EvidenceOrderField]

	EvidenceConnection struct {
		TotalCount int
		Edges      []*EvidenceEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewEvidenceConnection(
	p *page.Page[*coredata.Evidence, coredata.EvidenceOrderField],
	parentType any,
	parentID gid.GID,
) *EvidenceConnection {
	var edges = make([]*EvidenceEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewEvidenceEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &EvidenceConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewEvidenceEdge(e *coredata.Evidence, orderBy coredata.EvidenceOrderField) *EvidenceEdge {
	return &EvidenceEdge{
		Cursor: e.CursorKey(orderBy),
		Node:   NewEvidence(e),
	}
}

func NewEvidence(e *coredata.Evidence) *Evidence {
	var urlPtr *string = nil
	var fileObj *File = nil
	if e.URL != "" {
		urlCopy := e.URL
		urlPtr = &urlCopy
	}

	return &Evidence{
		ID:          e.ID,
		State:       e.State,
		Type:        e.Type,
		URL:         urlPtr,
		File:        fileObj,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
