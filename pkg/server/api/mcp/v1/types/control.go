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

func NewControl(c *coredata.Control) *Control {
	return &Control{
		ID:                          c.ID,
		OrganizationID:              c.OrganizationID,
		SectionTitle:                c.SectionTitle,
		FrameworkID:                 c.FrameworkID,
		Name:                        c.Name,
		Description:                 c.Description,
		BestPractice:                c.BestPractice,
		NotImplementedJustification: c.NotImplementedJustification,
		MaturityLevel:               ControlMaturityLevel(c.MaturityLevel),
		CreatedAt:                   c.CreatedAt,
		UpdatedAt:                   c.UpdatedAt,
	}
}

func NewListMeasureControlsOutput(controlPage *page.Page[*coredata.Control, coredata.ControlOrderField]) ListMeasureControlsOutput {
	controls := make([]*Control, 0, len(controlPage.Data))
	for _, c := range controlPage.Data {
		controls = append(controls, NewControl(c))
	}

	var nextCursor *page.CursorKey

	if len(controlPage.Data) > 0 {
		cursorKey := controlPage.Data[len(controlPage.Data)-1].CursorKey(controlPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListMeasureControlsOutput{
		NextCursor: nextCursor,
		Controls:   controls,
	}
}

func NewListControlsOutput(controlPage *page.Page[*coredata.Control, coredata.ControlOrderField]) ListControlsOutput {
	controls := make([]*Control, 0, len(controlPage.Data))
	for _, c := range controlPage.Data {
		controls = append(controls, NewControl(c))
	}

	var nextCursor *page.CursorKey

	if len(controlPage.Data) > 0 {
		cursorKey := controlPage.Data[len(controlPage.Data)-1].CursorKey(controlPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListControlsOutput{
		NextCursor: nextCursor,
		Controls:   controls,
	}
}
