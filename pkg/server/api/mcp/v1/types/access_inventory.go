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

func NewAccessInventoryGrant(g *coredata.AccessInventoryGrant) *AccessInventoryGrant {
	return &AccessInventoryGrant{
		ID:                      g.ID,
		OrganizationID:          g.OrganizationID,
		AccessReviewSourceID:    g.AccessReviewSourceID,
		PrincipalID:             g.PrincipalID,
		ResourceID:              g.ResourceID,
		PrincipalKey:            g.PrincipalKey,
		ResourceKey:             g.ResourceKey,
		EntitlementKey:          g.EntitlementKey,
		EntitlementDisplayName:  g.EntitlementDisplayName,
		Presence:                g.Presence,
		ConsecutiveMissingSyncs: g.ConsecutiveMissingSyncs,
		FirstSeenAt:             g.FirstSeenAt,
		LastSeenAt:              g.LastSeenAt,
		ProviderChangedAt:       g.ProviderChangedAt,
		CreatedAt:               g.CreatedAt,
		UpdatedAt:               g.UpdatedAt,
	}
}

func NewAccessInventoryPrincipal(p *coredata.AccessInventoryPrincipal) *AccessInventoryPrincipal {
	return &AccessInventoryPrincipal{
		ID:                   p.ID,
		OrganizationID:       p.OrganizationID,
		AccessReviewSourceID: p.AccessReviewSourceID,
		PrincipalKey:         p.PrincipalKey,
		Email:                p.Email,
		FullName:             p.FullName,
		ExternalID:           p.ExternalID,
		JobTitle:             p.JobTitle,
		IsAdmin:              p.IsAdmin,
		MfaStatus:            p.MFAStatus,
		AuthMethod:           p.AuthMethod,
		AccountType:          p.AccountType,
		Active:               p.Active,
		LastLogin:            p.LastLogin,
		AccountCreatedAt:     p.AccountCreatedAt,
		FirstSeenAt:          p.FirstSeenAt,
		LastSeenAt:           p.LastSeenAt,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

func NewAccessInventoryResource(r *coredata.AccessInventoryResource) *AccessInventoryResource {
	return &AccessInventoryResource{
		ID:                   r.ID,
		OrganizationID:       r.OrganizationID,
		AccessReviewSourceID: r.AccessReviewSourceID,
		ResourceKey:          r.ResourceKey,
		DisplayName:          r.DisplayName,
		Kind:                 r.Kind,
		FirstSeenAt:          r.FirstSeenAt,
		LastSeenAt:           r.LastSeenAt,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}

func NewListAccessInventoryGrantsOutput(
	p *page.Page[*coredata.AccessInventoryGrant, coredata.AccessInventoryGrantOrderField],
) ListAccessInventoryGrantsOutput {
	grants := make([]*AccessInventoryGrant, 0, len(p.Data))
	for _, g := range p.Data {
		grants = append(grants, NewAccessInventoryGrant(g))
	}

	var nextCursor *page.CursorKey
	if p.Info.HasNext && len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAccessInventoryGrantsOutput{
		NextCursor: nextCursor,
		Grants:     grants,
	}
}

func NewListAccessInventoryPrincipalsOutput(
	p *page.Page[*coredata.AccessInventoryPrincipal, coredata.AccessInventoryPrincipalOrderField],
) ListAccessInventoryPrincipalsOutput {
	principals := make([]*AccessInventoryPrincipal, 0, len(p.Data))
	for _, principal := range p.Data {
		principals = append(principals, NewAccessInventoryPrincipal(principal))
	}

	var nextCursor *page.CursorKey
	if p.Info.HasNext && len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAccessInventoryPrincipalsOutput{
		NextCursor: nextCursor,
		Principals: principals,
	}
}

func NewListAccessInventoryResourcesOutput(
	p *page.Page[*coredata.AccessInventoryResource, coredata.AccessInventoryResourceOrderField],
) ListAccessInventoryResourcesOutput {
	resources := make([]*AccessInventoryResource, 0, len(p.Data))
	for _, resource := range p.Data {
		resources = append(resources, NewAccessInventoryResource(resource))
	}

	var nextCursor *page.CursorKey
	if p.Info.HasNext && len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAccessInventoryResourcesOutput{
		NextCursor: nextCursor,
		Resources:  resources,
	}
}
