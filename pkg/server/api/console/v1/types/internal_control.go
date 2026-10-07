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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	InternalControlOrderBy OrderBy[coredata.InternalControlOrderField]

	InternalControlConnection struct {
		TotalCount int
		Edges      []*InternalControlEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.InternalControlFilter
		AsOf     *time.Time
	}
)

func NewInternalControlConnection(
	p *page.Page[*coredata.InternalControl, coredata.InternalControlOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.InternalControlFilter,
) *InternalControlConnection {
	var edges = make([]*InternalControlEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewInternalControlEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &InternalControlConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filters:  filters,
	}
}

func NewInternalControlConnectionAsOf(
	p *page.Page[*coredata.InternalControl, coredata.InternalControlOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.InternalControlFilter,
	asOf time.Time,
	totalCount int,
) *InternalControlConnection {
	edges := make([]*InternalControlEdge, len(p.Data))
	for i := range edges {
		edges[i] = &InternalControlEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewInternalControlAsOf(p.Data[i], asOf),
		}
	}

	return &InternalControlConnection{
		Edges:      edges,
		PageInfo:   *NewPageInfo(p),
		Resolver:   parentType,
		ParentID:   parentID,
		Filters:    filters,
		AsOf:       new(asOf),
		TotalCount: totalCount,
	}
}

func NewInternalControlEdge(c *coredata.InternalControl, orderBy coredata.InternalControlOrderField) *InternalControlEdge {
	return &InternalControlEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewInternalControl(c),
	}
}

func NewInternalControl(c *coredata.InternalControl) *InternalControl {
	internalControl := &InternalControl{
		ID:                   c.ID,
		Category:             c.Category,
		Name:                 c.Name,
		Description:          c.Description,
		State:                c.State,
		Code:                 c.Code,
		ControlType:          c.ControlType,
		Nature:               c.Nature,
		OperatingFrequency:   newOperatingFrequency(c),
		EvidenceCadence:      c.EvidenceCadence,
		TestingCadence:       c.TestingCadence,
		NextEvidenceDue:      c.NextEvidenceDue,
		NextTestDue:          c.NextTestDue,
		ImplementationStatus: c.ImplementationStatus,
		CreatedAt:            c.CreatedAt,
		UpdatedAt:            c.UpdatedAt,
	}

	if c.OwnerID != nil {
		internalControl.Owner = &Profile{ID: *c.OwnerID}
	}

	if c.ReviewerID != nil {
		internalControl.Reviewer = &Profile{ID: *c.ReviewerID}
	}

	return internalControl
}

func newOperatingFrequency(c *coredata.InternalControl) *coredata.InternalControlOperatingFrequency {
	if c.OperatingMode == nil {
		return nil
	}

	return &coredata.InternalControlOperatingFrequency{
		Mode:     *c.OperatingMode,
		Interval: c.OperatingInterval,
		Event:    c.OperatingEvent,
	}
}

func NewInternalControlAsOf(c *coredata.InternalControl, asOf time.Time) *InternalControl {
	internalControl := NewInternalControl(c)
	internalControl.AsOf = new(asOf)

	return internalControl
}
