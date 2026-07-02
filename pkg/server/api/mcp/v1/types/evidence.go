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

func NewEvidence(e *coredata.Evidence) *Evidence {
	return &Evidence{
		ID:             e.ID,
		OrganizationID: e.OrganizationID,
		MeasureID:      e.MeasureID,
		TaskID:         e.TaskID,
		State:          EvidenceState(e.State.String()),
		ReferenceID:    e.ReferenceID,
		Type:           EvidenceType(e.Type.String()),
		URL:            e.URL,
		Description:    e.Description,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func NewListMeasureEvidencesOutput(evidencePage *page.Page[*coredata.Evidence, coredata.EvidenceOrderField]) ListMeasureEvidencesOutput {
	evidences := make([]*Evidence, 0, len(evidencePage.Data))
	for _, v := range evidencePage.Data {
		evidences = append(evidences, NewEvidence(v))
	}

	var nextCursor *page.CursorKey

	if len(evidencePage.Data) > 0 {
		cursorKey := evidencePage.Data[len(evidencePage.Data)-1].CursorKey(evidencePage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListMeasureEvidencesOutput{
		NextCursor: nextCursor,
		Evidences:  evidences,
	}
}
