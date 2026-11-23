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

type ThirdPartyAgreementOrderBy OrderBy[coredata.ThirdPartyAgreementOrderField]

func NewThirdPartyAgreementConnection(
	p *page.Page[*coredata.ThirdPartyAgreement, coredata.ThirdPartyAgreementOrderField],
) *ThirdPartyAgreementConnection {
	edges := make([]*ThirdPartyAgreementEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewThirdPartyAgreementEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ThirdPartyAgreementConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewThirdPartyAgreementEdge(
	agreement *coredata.ThirdPartyAgreement,
	orderBy coredata.ThirdPartyAgreementOrderField,
) *ThirdPartyAgreementEdge {
	return &ThirdPartyAgreementEdge{
		Cursor: agreement.CursorKey(orderBy),
		Node:   NewThirdPartyAgreement(agreement),
	}
}

func NewThirdPartyAgreement(agreement *coredata.ThirdPartyAgreement) *ThirdPartyAgreement {
	return &ThirdPartyAgreement{
		ID: agreement.ID,
		ThirdParty: &ThirdParty{
			ID: agreement.ThirdPartyID,
		},
		Title:     agreement.Title,
		Type:      agreement.Type,
		Validity:  NewPeriod(agreement.ValidFrom, agreement.ValidUntil),
		File:      &File{ID: agreement.FileID},
		CreatedAt: agreement.CreatedAt,
		UpdatedAt: agreement.UpdatedAt,
	}
}
