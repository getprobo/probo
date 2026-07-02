// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
