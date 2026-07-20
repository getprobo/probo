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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ThirdPartyOrderBy OrderBy[coredata.ThirdPartyOrderField]

	ThirdPartyConnection struct {
		TotalCount int
		Edges      []*ThirdPartyEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.ThirdPartyFilter
	}
)

func NewThirdPartyConnection(
	p *page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.ThirdPartyFilter,
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
		Filters:  filters,
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
		ShowOnCompliancePortal:        v.ShowOnCompliancePortal,
		Level:                         v.Level,
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

	if v.ParentThirdPartyID != nil {
		object.ParentThirdParty = &ThirdParty{
			ID: *v.ParentThirdPartyID,
		}
	}

	return object
}
