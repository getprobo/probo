// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/probo"
)

type (
	ThirdPartyOrderBy OrderBy[coredata.ThirdPartyOrderField]

	ThirdPartyConnection struct {
		TotalCount int
		Edges      []*ThirdPartyEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewThirdPartyConnection(
	p *page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField],
	parentType any,
	parentID gid.GID,
) *ThirdPartyConnection {
	var edges = make([]*ThirdPartyEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewThirdPartyEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ThirdPartyConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewThirdPartyEdge(v *coredata.ThirdParty, orderBy coredata.ThirdPartyOrderField) *ThirdPartyEdge {
	return &ThirdPartyEdge{
		Cursor: v.CursorKey(orderBy),
		Node:   NewThirdParty(v),
	}
}

func NewThirdParty(v *coredata.ThirdParty) *ThirdParty {
	object := &ThirdParty{
		ID: v.ID,
		Organization: &Organization{
			ID: v.OrganizationID,
		},
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
		SnapshotID:                    v.SnapshotID,
		Countries:                     v.Countries,
		UpdatedAt:                     v.UpdatedAt,
		CreatedAt:                     v.CreatedAt,
	}

	if v.BusinessOwnerID != nil {
		object.BusinessOwner = &Profile{
			ID: *v.BusinessOwnerID,
		}
	}

	if v.SecurityOwnerID != nil {
		object.SecurityOwner = &Profile{
			ID: *v.SecurityOwnerID,
		}
	}

	return object
}

func NewThirdPartySubprocessors(sps []probo.Subprocessor) []*ThirdPartySubprocessor {
	result := make([]*ThirdPartySubprocessor, len(sps))
	for i, sp := range sps {
		result[i] = &ThirdPartySubprocessor{
			Name:    sp.Name,
			Country: sp.Country,
			Purpose: sp.Purpose,
		}
	}
	return result
}
