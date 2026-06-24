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
