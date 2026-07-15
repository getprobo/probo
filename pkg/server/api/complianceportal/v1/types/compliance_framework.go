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
	"go.probo.inc/probo/pkg/page"
)

type ComplianceFrameworkConnection struct {
	Edges    []*ComplianceFrameworkEdge `json:"edges"`
	PageInfo *PageInfo                  `json:"pageInfo"`
}

func NewComplianceFramework(cf *coredata.ComplianceFramework) *ComplianceFramework {
	return &ComplianceFramework{
		ID: cf.ID,
	}
}

func NewComplianceFrameworkEdge(cf *coredata.ComplianceFramework) *ComplianceFrameworkEdge {
	return &ComplianceFrameworkEdge{
		Cursor: cf.CursorKey(coredata.ComplianceFrameworkOrderFieldRank),
		Node:   NewComplianceFramework(cf),
	}
}

func NewComplianceFrameworkConnection(
	p *page.Page[*coredata.ComplianceFramework, coredata.ComplianceFrameworkOrderField],
) *ComplianceFrameworkConnection {
	edges := make([]*ComplianceFrameworkEdge, len(p.Data))
	for i, cf := range p.Data {
		edges[i] = NewComplianceFrameworkEdge(cf)
	}

	return &ComplianceFrameworkConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}
