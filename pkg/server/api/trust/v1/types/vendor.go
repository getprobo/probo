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
	SubprocessorConnection struct {
		TotalCount int
		Edges      []*SubprocessorEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewSubprocessorConnection(
	p *page.Page[*coredata.Vendor, coredata.VendorOrderField],
	parentType any,
	parentID gid.GID,
) *SubprocessorConnection {
	edges := make([]*SubprocessorEdge, len(p.Data))
	for i, vendor := range p.Data {
		edges[i] = NewSubprocessorEdge(vendor, p.Cursor.OrderBy.Field)
	}

	return &SubprocessorConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewSubprocessor(v *coredata.Vendor) *Subprocessor {
	return &Subprocessor{
		ID:               v.ID,
		Name:             v.Name,
		Description:      v.Description,
		Category:         v.Category,
		WebsiteURL:       v.WebsiteURL,
		PrivacyPolicyURL: v.PrivacyPolicyURL,
		Countries:        []coredata.CountryCode(v.Countries),
	}
}

func NewSubprocessorEdge(v *coredata.Vendor, orderField coredata.VendorOrderField) *SubprocessorEdge {
	return &SubprocessorEdge{
		Node:   NewSubprocessor(v),
		Cursor: v.CursorKey(orderField),
	}
}
