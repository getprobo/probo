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

func NewObligation(o *coredata.Obligation) *Obligation {
	obligation := &Obligation{
		ID:                     o.ID,
		OrganizationID:         o.OrganizationID,
		Area:                   o.Area,
		Source:                 o.Source,
		Requirement:            o.Requirement,
		ActionsToBeImplemented: o.ActionsToBeImplemented,
		Regulator:              o.Regulator,
		OwnerID:                o.OwnerID,
		LastReviewDate:         o.LastReviewDate,
		DueDate:                o.DueDate,
		Status:                 o.Status,
		Type:                   o.Type,
		CreatedAt:              o.CreatedAt,
		UpdatedAt:              o.UpdatedAt,
	}

	return obligation
}

func NewListObligationsOutput(obligationPage *page.Page[*coredata.Obligation, coredata.ObligationOrderField]) ListObligationsOutput {
	obligations := make([]*Obligation, 0, len(obligationPage.Data))
	for _, v := range obligationPage.Data {
		obligations = append(obligations, NewObligation(v))
	}

	var nextCursor *page.CursorKey

	if len(obligationPage.Data) > 0 {
		cursorKey := obligationPage.Data[len(obligationPage.Data)-1].CursorKey(obligationPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListObligationsOutput{
		NextCursor:  nextCursor,
		Obligations: obligations,
	}
}

func NewListControlObligationsOutput(obligationPage *page.Page[*coredata.Obligation, coredata.ObligationOrderField]) ListControlObligationsOutput {
	obligations := make([]*Obligation, 0, len(obligationPage.Data))
	for _, v := range obligationPage.Data {
		obligations = append(obligations, NewObligation(v))
	}

	var nextCursor *page.CursorKey

	if len(obligationPage.Data) > 0 {
		cursorKey := obligationPage.Data[len(obligationPage.Data)-1].CursorKey(obligationPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListControlObligationsOutput{
		NextCursor:  nextCursor,
		Obligations: obligations,
	}
}

func NewListRiskObligationsOutput(obligationPage *page.Page[*coredata.Obligation, coredata.ObligationOrderField]) ListRiskObligationsOutput {
	obligations := make([]*Obligation, 0, len(obligationPage.Data))
	for _, v := range obligationPage.Data {
		obligations = append(obligations, NewObligation(v))
	}

	var nextCursor *page.CursorKey

	if len(obligationPage.Data) > 0 {
		cursorKey := obligationPage.Data[len(obligationPage.Data)-1].CursorKey(obligationPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskObligationsOutput{
		NextCursor:  nextCursor,
		Obligations: obligations,
	}
}
