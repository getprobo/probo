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
	AgentRunOrderBy OrderBy[coredata.AgentRunOrderField]

	AgentRunConnection struct {
		TotalCount int
		Edges      []*AgentRunEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewAgentRunConnection(
	p *page.Page[*coredata.AgentRun, coredata.AgentRunOrderField],
	parentType any,
	parentID gid.GID,
) *AgentRunConnection {
	var edges = make([]*AgentRunEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAgentRunEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AgentRunConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAgentRunEdge(run *coredata.AgentRun, orderBy coredata.AgentRunOrderField) *AgentRunEdge {
	return &AgentRunEdge{
		Cursor: run.CursorKey(orderBy),
		Node:   NewAgentRun(run),
	}
}

func NewAgentRun(run *coredata.AgentRun) *AgentRun {
	return &AgentRun{
		ID: run.ID,
		Organization: &Organization{
			ID: run.OrganizationID,
		},
		AgentName:    run.StartAgentName,
		Status:       run.Status,
		ErrorMessage: run.ErrorMessage,
		StartedAt:    run.StartedAt,
		CreatedAt:    run.CreatedAt,
		UpdatedAt:    run.UpdatedAt,
	}
}
