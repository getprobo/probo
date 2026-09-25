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
	AgentExecutionOrderBy OrderBy[coredata.AgentExecutionOrderField]

	AgentExecutionConnection struct {
		TotalCount int
		Edges      []*AgentExecutionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewAgentExecutionConnection(
	p *page.Page[*coredata.AgentExecution, coredata.AgentExecutionOrderField],
	parentType any,
	parentID gid.GID,
) *AgentExecutionConnection {
	var edges = make([]*AgentExecutionEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAgentExecutionEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AgentExecutionConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAgentExecutionEdge(run *coredata.AgentExecution, orderBy coredata.AgentExecutionOrderField) *AgentExecutionEdge {
	return &AgentExecutionEdge{
		Cursor: run.CursorKey(orderBy),
		Node:   NewAgentExecution(run),
	}
}

func NewAgentExecution(run *coredata.AgentExecution) *AgentExecution {
	return &AgentExecution{
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
