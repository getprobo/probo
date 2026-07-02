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
	DocumentVersionApprovalDecisionOrderBy OrderBy[coredata.DocumentVersionApprovalDecisionOrderField]

	DocumentVersionApprovalDecisionConnection struct {
		TotalCount int
		Edges      []*DocumentVersionApprovalDecisionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.DocumentVersionApprovalDecisionFilter
	}
)

func NewDocumentVersionApprovalDecisionConnection(
	page *page.Page[*coredata.DocumentVersionApprovalDecision, coredata.DocumentVersionApprovalDecisionOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.DocumentVersionApprovalDecisionFilter,
) *DocumentVersionApprovalDecisionConnection {
	edges := make([]*DocumentVersionApprovalDecisionEdge, len(page.Data))
	for i, decision := range page.Data {
		edges[i] = NewDocumentVersionApprovalDecisionEdge(decision, page.Cursor.OrderBy.Field)
	}

	return &DocumentVersionApprovalDecisionConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(page),

		Resolver: parentType,
		ParentID: parentID,
		Filters:  filter,
	}
}

func NewDocumentVersionApprovalDecisionEdge(decision *coredata.DocumentVersionApprovalDecision, orderBy coredata.DocumentVersionApprovalDecisionOrderField) *DocumentVersionApprovalDecisionEdge {
	return &DocumentVersionApprovalDecisionEdge{
		Cursor: decision.CursorKey(orderBy),
		Node:   NewDocumentVersionApprovalDecision(decision),
	}
}

func NewDocumentVersionApprovalDecision(decision *coredata.DocumentVersionApprovalDecision) *DocumentVersionApprovalDecision {
	return &DocumentVersionApprovalDecision{
		Quorum: &DocumentVersionApprovalQuorum{
			ID: decision.QuorumID,
		},
		Approver: &Profile{
			ID: decision.ApproverID,
		},
		ID:        decision.ID,
		State:     decision.State,
		Comment:   decision.Comment,
		DecidedAt: decision.DecidedAt,
		CreatedAt: decision.CreatedAt,
		UpdatedAt: decision.UpdatedAt,
	}
}
