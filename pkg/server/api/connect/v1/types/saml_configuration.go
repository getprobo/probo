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
	SAMLConfigurationOrderBy OrderBy[coredata.SAMLConfigurationOrderField]

	SAMLConfigurationConnection struct {
		TotalCount int
		Edges      []*SAMLConfigurationEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewSAMLConfigurationConnection(
	p *page.Page[*coredata.SAMLConfiguration, coredata.SAMLConfigurationOrderField],
	resolver any,
	parentID gid.GID,
) *SAMLConfigurationConnection {
	edges := make([]*SAMLConfigurationEdge, len(p.Data))
	for i, samlConfiguration := range p.Data {
		edges[i] = NewSAMLConfigurationEdge(samlConfiguration, p.Cursor.OrderBy.Field)
	}

	return &SAMLConfigurationConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewSAMLConfigurationEdge(samlConfiguration *coredata.SAMLConfiguration, orderField coredata.SAMLConfigurationOrderField) *SAMLConfigurationEdge {
	return &SAMLConfigurationEdge{
		Node:   NewSAMLConfiguration(samlConfiguration),
		Cursor: samlConfiguration.CursorKey(orderField),
	}
}

func NewSAMLConfiguration(samlConfiguration *coredata.SAMLConfiguration) *SAMLConfiguration {
	return &SAMLConfiguration{
		ID:                      samlConfiguration.ID,
		EmailDomain:             samlConfiguration.EmailDomain,
		EnforcementPolicy:       samlConfiguration.EnforcementPolicy,
		DomainVerifiedAt:        samlConfiguration.DomainVerifiedAt,
		DomainVerificationToken: samlConfiguration.DomainVerificationToken,
		IdpEntityID:             samlConfiguration.IdPEntityID,
		IdpSsoURL:               samlConfiguration.IdPSsoURL,
		IdpCertificate:          samlConfiguration.IdPCertificate,
		AutoSignupEnabled:       samlConfiguration.AutoSignupEnabled,
		CreatedAt:               samlConfiguration.CreatedAt,
		UpdatedAt:               samlConfiguration.UpdatedAt,
		AttributeMappings: &SAMLAttributeMappings{
			Email:     samlConfiguration.AttributeEmail,
			FirstName: samlConfiguration.AttributeFirstname,
			LastName:  samlConfiguration.AttributeLastname,
			Role:      samlConfiguration.AttributeRole,
		},
	}
}
