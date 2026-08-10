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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AiSystemOrderBy OrderBy[coredata.AiSystemOrderField]

	AiSystemConnection struct {
		TotalCount int
		Edges      []*AiSystemEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *AiSystemFilter
	}
)

func NewAiSystemConnection(
	p *page.Page[*coredata.AiSystem, coredata.AiSystemOrderField],
	parentType any,
	parentID gid.GID,
	filter *AiSystemFilter,
) *AiSystemConnection {
	edges := make([]*AiSystemEdge, len(p.Data))
	for i, aiSystem := range p.Data {
		edges[i] = NewAiSystemEdge(aiSystem, p.Cursor.OrderBy.Field)
	}

	return &AiSystemConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewAiSystemEdge(
	s *coredata.AiSystem,
	orderField coredata.AiSystemOrderField,
) *AiSystemEdge {
	return &AiSystemEdge{
		Node:   NewAiSystem(s),
		Cursor: s.CursorKey(orderField),
	}
}

func NewAiSystem(s *coredata.AiSystem) *AiSystem {
	riskClassification := coredata.AiSystemRiskClassificationMinimal
	if s.RiskClassification != nil {
		riskClassification = *s.RiskClassification
	}

	aiSystem := &AiSystem{
		ID: s.ID,
		Organization: &Organization{
			ID: s.OrganizationID,
		},
		Name:                    s.Name,
		Version:                 s.Version,
		CompanyRoles:            s.CompanyRoles,
		Status:                  s.Status,
		Source:                  s.Source,
		Purpose:                 s.Purpose,
		IntendedUseCases:        s.IntendedUseCases,
		AutonomyLevel:           s.AutonomyLevel,
		HumanOversightMechanism: s.HumanOversightMechanism,
		RiskClassification:      riskClassification,
		KeyStakeholders:         s.KeyStakeholders,
		DataSourcesAndType:      s.DataSourcesAndType,
		DeploymentDate:          s.DeploymentDate,
		LastReviewDate:          s.LastReviewDate,
		NextReviewDate:          s.NextReviewDate,
		Notes:                   s.Notes,
		CreatedAt:               s.CreatedAt,
		UpdatedAt:               s.UpdatedAt,
	}

	if s.OwnerID != nil {
		aiSystem.Owner = &Profile{
			ID: *s.OwnerID,
		}
	}

	return aiSystem
}
