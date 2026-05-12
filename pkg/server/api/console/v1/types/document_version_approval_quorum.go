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
	DocumentVersionApprovalQuorumConnection struct {
		TotalCount int
		Edges      []*DocumentVersionApprovalQuorumEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewDocumentVersionApprovalQuorumConnection(
	page *page.Page[*coredata.DocumentVersionApprovalQuorum, coredata.DocumentVersionApprovalQuorumOrderField],
	parentType any,
	parentID gid.GID,
) *DocumentVersionApprovalQuorumConnection {
	edges := make([]*DocumentVersionApprovalQuorumEdge, len(page.Data))
	for i, quorum := range page.Data {
		edges[i] = NewDocumentVersionApprovalQuorumEdge(quorum, page.Cursor.OrderBy.Field)
	}

	return &DocumentVersionApprovalQuorumConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(page),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewDocumentVersionApprovalQuorumEdge(
	quorum *coredata.DocumentVersionApprovalQuorum,
	orderBy coredata.DocumentVersionApprovalQuorumOrderField,
) *DocumentVersionApprovalQuorumEdge {
	return &DocumentVersionApprovalQuorumEdge{
		Cursor: quorum.CursorKey(orderBy),
		Node:   NewDocumentVersionApprovalQuorum(quorum),
	}
}

func NewDocumentVersionApprovalQuorum(quorum *coredata.DocumentVersionApprovalQuorum) *DocumentVersionApprovalQuorum {
	return &DocumentVersionApprovalQuorum{
		ID: quorum.ID,
		DocumentVersion: &DocumentVersion{
			ID: quorum.VersionID,
		},
		Status:    quorum.Status,
		CreatedAt: quorum.CreatedAt,
		UpdatedAt: quorum.UpdatedAt,
	}
}
