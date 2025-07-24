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
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
)

type (
	VendorOrderBy OrderBy[coredata.VendorOrderField]

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
	var edges = make([]*VendorEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewVendorEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &VendorConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewVendorEdge(v *coredata.Vendor, orderBy coredata.VendorOrderField) *VendorEdge {
	return &VendorEdge{
		Cursor: v.CursorKey(orderBy),
		Node:   NewVendor(v),
	}
}

func NewVendor(v *coredata.Vendor) *Vendor {
	return &Vendor{
		ID:                            v.ID,
		Name:                          v.Name,
		Description:                   v.Description,
		StatusPageURL:                 v.StatusPageURL,
		TermsOfServiceURL:             v.TermsOfServiceURL,
		PrivacyPolicyURL:              v.PrivacyPolicyURL,
		ServiceLevelAgreementURL:      v.ServiceLevelAgreementURL,
		DataProcessingAgreementURL:    v.DataProcessingAgreementURL,
		BusinessAssociateAgreementURL: v.BusinessAssociateAgreementURL,
		SubprocessorsListURL:          v.SubprocessorsListURL,
		Certifications:                v.Certifications,
		SecurityPageURL:               v.SecurityPageURL,
		TrustPageURL:                  v.TrustPageURL,
		HeadquarterAddress:            v.HeadquarterAddress,
		LegalName:                     v.LegalName,
		WebsiteURL:                    v.WebsiteURL,
		Category:                      v.Category,
		ShowOnTrustCenter:             v.ShowOnTrustCenter,
		UpdatedAt:                     v.UpdatedAt,
		CreatedAt:                     v.CreatedAt,
	}
}
