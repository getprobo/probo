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
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CommonGVLVendorOrderBy OrderBy[coredata.CommonGVLVendorOrderField]

	CommonGVLVendorFilter struct {
		Query          *string
		CookieBannerID *gid.GID
		Membership     *coredata.CommonGVLVendorMembership
	}

	CommonGVLVendorConnection struct {
		TotalCount int
		Edges      []*CommonGVLVendorEdge
		PageInfo   PageInfo

		Resolver any
		ParentID *gid.GID
		Filter   *coredata.CommonGVLVendorFilter
	}

	CommonGVLCatalog struct {
		VendorListVersion *int
		TcfPolicyVersion  *int
	}
)

func NewCommonGVLCatalog(c *cookiebanner.CommonGVLCatalog) *CommonGVLCatalog {
	if c == nil {
		return &CommonGVLCatalog{}
	}

	return &CommonGVLCatalog{
		VendorListVersion: c.VendorListVersion,
		TcfPolicyVersion:  c.TCFPolicyVersion,
	}
}

func NewCommonGVLVendor(v *coredata.CommonGVLVendor) *CommonGVLVendor {
	return &CommonGVLVendor{
		ID:              v.ID,
		IabVendorID:     v.IABVendorID,
		Name:            v.Name,
		PolicyURL:       v.PolicyURL,
		Purposes:        int32sToInts(v.Purposes),
		LegIntPurposes:  int32sToInts(v.LegIntPurposes),
		SpecialFeatures: int32sToInts(v.SpecialFeatures),
	}
}

func NewCommonGVLVendorConnection(
	p *page.Page[*coredata.CommonGVLVendor, coredata.CommonGVLVendorOrderField],
	parentType any,
	parentID *gid.GID,
	filter *coredata.CommonGVLVendorFilter,
) *CommonGVLVendorConnection {
	edges := make([]*CommonGVLVendorEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCommonGVLVendorEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CommonGVLVendorConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewCommonGVLVendorEdge(v *coredata.CommonGVLVendor, orderBy coredata.CommonGVLVendorOrderField) *CommonGVLVendorEdge {
	return &CommonGVLVendorEdge{
		Cursor: v.CursorKey(orderBy),
		Node:   NewCommonGVLVendor(v),
	}
}

func int32sToInts(values []int32) []int {
	if values == nil {
		return []int{}
	}

	out := make([]int, len(values))
	for i, v := range values {
		out[i] = int(v)
	}

	return out
}
