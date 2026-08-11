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
	BusinessFunctionOrderBy OrderBy[coredata.BusinessFunctionOrderField]

	BusinessFunctionConnection struct {
		TotalCount int
		Edges      []*BusinessFunctionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *BusinessFunctionFilter
	}
)

func NewBusinessFunctionConnection(
	p *page.Page[*coredata.BusinessFunction, coredata.BusinessFunctionOrderField],
	parentType any,
	parentID gid.GID,
	filter *BusinessFunctionFilter,
) *BusinessFunctionConnection {
	edges := make([]*BusinessFunctionEdge, len(p.Data))
	for i, businessFunction := range p.Data {
		edges[i] = NewBusinessFunctionEdge(businessFunction, p.Cursor.OrderBy.Field)
	}

	return &BusinessFunctionConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewBusinessFunctionEdge(
	bf *coredata.BusinessFunction,
	orderField coredata.BusinessFunctionOrderField,
) *BusinessFunctionEdge {
	return &BusinessFunctionEdge{
		Node:   NewBusinessFunction(bf),
		Cursor: bf.CursorKey(orderField),
	}
}

func NewBusinessFunction(bf *coredata.BusinessFunction) *BusinessFunction {
	businessFunction := &BusinessFunction{
		ID: bf.ID,
		Organization: &Organization{
			ID: bf.OrganizationID,
		},
		ReferenceID:     bf.ReferenceID,
		Name:            bf.Name,
		Classification:  bf.Classification,
		MtdMinutes:      bf.MTDMinutes,
		RtoMinutes:      bf.RTOMinutes,
		RpoMinutes:      bf.RPOMinutes,
		ImpactTolerance: bf.ImpactTolerance,
		Notes:           bf.Notes,
		CreatedAt:       bf.CreatedAt,
		UpdatedAt:       bf.UpdatedAt,
	}

	if bf.OwnerID != nil {
		businessFunction.Owner = &Profile{
			ID: *bf.OwnerID,
		}
	}

	return businessFunction
}
