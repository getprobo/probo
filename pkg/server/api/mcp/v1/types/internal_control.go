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

func NewInternalControl(m *coredata.InternalControl) *InternalControl {
	return &InternalControl{
		ID:                   m.ID,
		Category:             m.Category,
		Name:                 m.Name,
		Description:          m.Description,
		State:                m.State,
		Code:                 m.Code,
		ControlType:          m.ControlType,
		Nature:               m.Nature,
		OperatingFrequency:   newOperatingFrequency(m),
		EvidenceCadence:      m.EvidenceCadence,
		TestingCadence:       m.TestingCadence,
		NextEvidenceDue:      m.NextEvidenceDue,
		NextTestDue:          m.NextTestDue,
		ImplementationStatus: m.ImplementationStatus,
		OwnerID:              m.OwnerID,
		ReviewerID:           m.ReviewerID,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

func NewListControlInternalControlsOutput(internalControlPage *page.Page[*coredata.InternalControl, coredata.InternalControlOrderField]) ListControlInternalControlsOutput {
	internalControls := make([]*InternalControl, 0, len(internalControlPage.Data))
	for _, v := range internalControlPage.Data {
		internalControls = append(internalControls, NewInternalControl(v))
	}

	var nextCursor *page.CursorKey

	if len(internalControlPage.Data) > 0 {
		cursorKey := internalControlPage.Data[len(internalControlPage.Data)-1].CursorKey(internalControlPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListControlInternalControlsOutput{
		NextCursor:       nextCursor,
		InternalControls: internalControls,
	}
}

func NewListRiskInternalControlsOutput(internalControlPage *page.Page[*coredata.InternalControl, coredata.InternalControlOrderField]) ListRiskInternalControlsOutput {
	internalControls := make([]*InternalControl, 0, len(internalControlPage.Data))
	for _, v := range internalControlPage.Data {
		internalControls = append(internalControls, NewInternalControl(v))
	}

	var nextCursor *page.CursorKey

	if len(internalControlPage.Data) > 0 {
		cursorKey := internalControlPage.Data[len(internalControlPage.Data)-1].CursorKey(internalControlPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskInternalControlsOutput{
		NextCursor:       nextCursor,
		InternalControls: internalControls,
	}
}

func NewListInternalControlsOutput(internalControlPage *page.Page[*coredata.InternalControl, coredata.InternalControlOrderField]) ListInternalControlsOutput {
	internalControls := make([]*InternalControl, 0, len(internalControlPage.Data))
	for _, v := range internalControlPage.Data {
		internalControls = append(internalControls, NewInternalControl(v))
	}

	var nextCursor *page.CursorKey

	if len(internalControlPage.Data) > 0 {
		cursorKey := internalControlPage.Data[len(internalControlPage.Data)-1].CursorKey(internalControlPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListInternalControlsOutput{
		NextCursor:       nextCursor,
		InternalControls: internalControls,
	}
}

func newOperatingFrequency(m *coredata.InternalControl) *coredata.InternalControlOperatingFrequency {
	if m.OperatingMode == nil {
		return nil
	}

	return &coredata.InternalControlOperatingFrequency{
		Mode:     *m.OperatingMode,
		Interval: m.OperatingInterval,
		Event:    m.OperatingEvent,
	}
}
