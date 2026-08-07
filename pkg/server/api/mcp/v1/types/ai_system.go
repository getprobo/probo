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

func NewAiSystem(s *coredata.AiSystem) *AiSystem {
	return &AiSystem{
		ID:                      s.ID,
		OrganizationID:          s.OrganizationID,
		Name:                    s.Name,
		Version:                 s.Version,
		CompanyRoles:            s.CompanyRoles,
		Status:                  s.Status,
		OwnerID:                 s.OwnerID,
		Source:                  s.Source,
		Purpose:                 s.Purpose,
		IntendedUseCases:        s.IntendedUseCases,
		AutonomyLevel:           s.AutonomyLevel,
		HumanOversightMechanism: s.HumanOversightMechanism,
		RiskClassification:      s.RiskClassification,
		KeyStakeholders:         s.KeyStakeholders,
		DataSourcesAndType:      s.DataSourcesAndType,
		DeploymentDate:          s.DeploymentDate,
		LastReviewDate:          s.LastReviewDate,
		NextReviewDate:          s.NextReviewDate,
		Notes:                   s.Notes,
		CreatedAt:               s.CreatedAt,
		UpdatedAt:               s.UpdatedAt,
	}
}

func NewListAiSystemsOutput(
	aiSystemPage *page.Page[*coredata.AiSystem, coredata.AiSystemOrderField],
) ListAiSystemsOutput {
	aiSystems := make([]*AiSystem, 0, len(aiSystemPage.Data))
	for _, v := range aiSystemPage.Data {
		aiSystems = append(aiSystems, NewAiSystem(v))
	}

	var nextCursor *page.CursorKey

	if len(aiSystemPage.Data) > 0 {
		cursorKey := aiSystemPage.Data[len(aiSystemPage.Data)-1].CursorKey(aiSystemPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAiSystemsOutput{
		NextCursor: nextCursor,
		AiSystems:  aiSystems,
	}
}
