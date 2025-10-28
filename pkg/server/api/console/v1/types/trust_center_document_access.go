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
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
)

type (
	TrustCenterDocumentAccessOrderBy = OrderBy[coredata.TrustCenterDocumentAccessOrderField]

	TrustCenterDocumentAccessConnection struct {
		TotalCount int
		Edges      []*TrustCenterDocumentAccessEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	TrustCenterDocumentAccess struct {
		ID                gid.GID            `json:"id"`
		Active            bool               `json:"active"`
		Requested         bool               `json:"requested"`
		CreatedAt         time.Time          `json:"createdAt"`
		UpdatedAt         time.Time          `json:"updatedAt"`
		TrustCenterAccess *TrustCenterAccess `json:"trustCenterAccess"`
		Document          *Document          `json:"document,omitempty"`
		Report            *Report            `json:"report,omitempty"`
		TrustCenterFile   *TrustCenterFile   `json:"trustCenterFile,omitempty"`

		// Internal fields used by resolvers
		DocumentID        *gid.GID `json:"-"`
		ReportID          *gid.GID `json:"-"`
		TrustCenterFileID *gid.GID `json:"-"`
	}
)

func NewTrustCenterDocumentAccess(tcda *coredata.TrustCenterDocumentAccess) *TrustCenterDocumentAccess {
	return &TrustCenterDocumentAccess{
		ID:                tcda.ID,
		Active:            tcda.Active,
		Requested:         tcda.Requested,
		CreatedAt:         tcda.CreatedAt,
		UpdatedAt:         tcda.UpdatedAt,
		DocumentID:        tcda.DocumentID,
		ReportID:          tcda.ReportID,
		TrustCenterFileID: tcda.TrustCenterFileID,
	}
}

func NewTrustCenterDocumentAccessConnection(
	p *page.Page[*coredata.TrustCenterDocumentAccess, coredata.TrustCenterDocumentAccessOrderField],
	parentType any,
	parentID gid.GID,
) *TrustCenterDocumentAccessConnection {
	var edges = make([]*TrustCenterDocumentAccessEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTrustCenterDocumentAccessEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TrustCenterDocumentAccessConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTrustCenterDocumentAccessEdges(accesses []*coredata.TrustCenterDocumentAccess, orderBy coredata.TrustCenterDocumentAccessOrderField) []*TrustCenterDocumentAccessEdge {
	edges := make([]*TrustCenterDocumentAccessEdge, len(accesses))

	for i := range edges {
		edges[i] = NewTrustCenterDocumentAccessEdge(accesses[i], orderBy)
	}

	return edges
}

func NewTrustCenterDocumentAccessEdge(access *coredata.TrustCenterDocumentAccess, orderBy coredata.TrustCenterDocumentAccessOrderField) *TrustCenterDocumentAccessEdge {
	return &TrustCenterDocumentAccessEdge{
		Cursor: access.CursorKey(orderBy),
		Node:   NewTrustCenterDocumentAccess(access),
	}
}
