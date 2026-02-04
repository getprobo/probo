// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	VendorConnection struct {
		TotalCount int
		Edges      []*VendorEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewVendorConnection(
	p *page.Page[*coredata.Vendor, coredata.VendorOrderField],
	parentType any,
	parentID gid.GID,
) *VendorConnection {
	edges := make([]*VendorEdge, len(p.Data))
	for i, vendor := range p.Data {
		edges[i] = NewVendorEdge(vendor, p.Cursor.OrderBy.Field)
	}

	return &VendorConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewVendor(v *coredata.Vendor) *Vendor {
	return &Vendor{
		ID:               v.ID,
		Name:             v.Name,
		Description:      v.Description,
		Category:         v.Category,
		WebsiteURL:       v.WebsiteURL,
		PrivacyPolicyURL: v.PrivacyPolicyURL,
		Countries:        []coredata.CountryCode(v.Countries),
	}
}

func NewVendorEdge(v *coredata.Vendor, orderField coredata.VendorOrderField) *VendorEdge {
	return &VendorEdge{
		Node:   NewVendor(v),
		Cursor: v.CursorKey(orderField),
	}
}
