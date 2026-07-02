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

func NewFinding(f *coredata.Finding) *Finding {
	finding := &Finding{
		ID:                 f.ID,
		OrganizationID:     f.OrganizationID,
		Kind:               f.Kind,
		ReferenceID:        f.ReferenceID,
		Description:        f.Description,
		Source:             f.Source,
		IdentifiedOn:       f.IdentifiedOn,
		RootCause:          f.RootCause,
		CorrectiveAction:   f.CorrectiveAction,
		OwnerID:            f.OwnerID,
		DueDate:            f.DueDate,
		Status:             f.Status,
		Priority:           f.Priority,
		RiskID:             f.RiskID,
		EffectivenessCheck: f.EffectivenessCheck,
		CreatedAt:          f.CreatedAt,
		UpdatedAt:          f.UpdatedAt,
	}

	return finding
}

func NewListFindingsOutput(findingPage *page.Page[*coredata.Finding, coredata.FindingOrderField]) ListFindingsOutput {
	findings := make([]*Finding, 0, len(findingPage.Data))
	for _, v := range findingPage.Data {
		findings = append(findings, NewFinding(v))
	}

	var nextCursor *page.CursorKey

	if len(findingPage.Data) > 0 {
		cursorKey := findingPage.Data[len(findingPage.Data)-1].CursorKey(findingPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListFindingsOutput{
		NextCursor: nextCursor,
		Findings:   findings,
	}
}
