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
