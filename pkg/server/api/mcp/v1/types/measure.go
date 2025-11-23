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

func NewMeasure(m *coredata.Measure) *Measure {
	return &Measure{
		ID:          m.ID,
		Category:    m.Category,
		Name:        m.Name,
		Description: m.Description,
		State:       m.State,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func NewListMeasuresOutput(measurePage *page.Page[*coredata.Measure, coredata.MeasureOrderField]) ListMeasuresOutput {
	measures := make([]*Measure, 0, len(measurePage.Data))
	for _, v := range measurePage.Data {
		measures = append(measures, NewMeasure(v))
	}

	var nextCursor *page.CursorKey
	if len(measurePage.Data) > 0 {
		cursorKey := measurePage.Data[len(measurePage.Data)-1].CursorKey(measurePage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListMeasuresOutput{
		NextCursor: nextCursor,
		Measures:   measures,
	}
}
