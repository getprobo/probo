// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	if e.URL != "" {
		urlCopy := e.URL
		urlPtr = &urlCopy
	}

	evidence := &Evidence{
		ID:    e.ID,
		State: e.State,
		Type:  e.Type,
		URL:   urlPtr,
		Measure: &Measure{
			ID: e.MeasureID,
		},
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	if e.EvidenceFileId != nil {
		evidence.File = &File{
			ID: *e.EvidenceFileId,
		}
	}

	if e.TaskID != nil {
		evidence.Task = &Task{
			ID: *e.TaskID,
		}
	}

	return evidence
}
