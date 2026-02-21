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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewControl(c *coredata.Control) *Control {
	return &Control{
		ID:             c.ID,
		OrganizationID: c.OrganizationID,
		SectionTitle:   c.SectionTitle,
		FrameworkID:    c.FrameworkID,
		Name:           c.Name,
		Description:    c.Description,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
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
