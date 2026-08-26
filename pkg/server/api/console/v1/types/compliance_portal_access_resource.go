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
	CompliancePortalAccessResourceOrderBy OrderBy[coredata.CompliancePortalAccessResourceOrderField]

	CompliancePortalAccessResourceConnection struct {
		TotalCount int
		Edges      []*CompliancePortalAccessResourceEdge
		PageInfo   *PageInfo

		ParentID gid.GID
	}
)

func NewCompliancePortalAccessResource(
	resource *coredata.CompliancePortalAccessResource,
) *CompliancePortalAccessResource {
	return &CompliancePortalAccessResource{
		Kind:       resource.Kind,
		ResourceID: resource.ID,
		Name:       resource.Name,
		Category:   resource.Category,
		Status:     resource.Status,
	}
}

func NewCompliancePortalAccessResourceConnection(
	p *page.Page[*coredata.CompliancePortalAccessResource, coredata.CompliancePortalAccessResourceOrderField],
	parentID gid.GID,
) *CompliancePortalAccessResourceConnection {
	edges := make([]*CompliancePortalAccessResourceEdge, len(p.Data))
	for i, resource := range p.Data {
		edges[i] = NewCompliancePortalAccessResourceEdge(resource, p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalAccessResourceConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		ParentID: parentID,
	}
}

func NewCompliancePortalAccessResourceEdge(
	resource *coredata.CompliancePortalAccessResource,
	orderBy coredata.CompliancePortalAccessResourceOrderField,
) *CompliancePortalAccessResourceEdge {
	return &CompliancePortalAccessResourceEdge{
		Cursor: resource.CursorKey(orderBy),
		Node:   NewCompliancePortalAccessResource(resource),
	}
}
