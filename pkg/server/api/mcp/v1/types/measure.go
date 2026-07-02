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

func NewListControlMeasuresOutput(measurePage *page.Page[*coredata.Measure, coredata.MeasureOrderField]) ListControlMeasuresOutput {
	measures := make([]*Measure, 0, len(measurePage.Data))
	for _, v := range measurePage.Data {
		measures = append(measures, NewMeasure(v))
	}

	var nextCursor *page.CursorKey

	if len(measurePage.Data) > 0 {
		cursorKey := measurePage.Data[len(measurePage.Data)-1].CursorKey(measurePage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListControlMeasuresOutput{
		NextCursor: nextCursor,
		Measures:   measures,
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
