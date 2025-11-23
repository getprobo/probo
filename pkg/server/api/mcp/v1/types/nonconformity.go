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

func NewNonconformity(n *coredata.Nonconformity) *Nonconformity {
	return &Nonconformity{
		ID:                 n.ID,
		SnapshotID:         n.SnapshotID,
		ReferenceID:        n.ReferenceID,
		Description:        n.Description,
		AuditID:            n.AuditID,
		DateIdentified:     n.DateIdentified,
		RootCause:          n.RootCause,
		CorrectiveAction:   n.CorrectiveAction,
		OwnerID:            n.OwnerID,
		DueDate:            n.DueDate,
		Status:             n.Status,
		EffectivenessCheck: n.EffectivenessCheck,
		CreatedAt:          n.CreatedAt,
		UpdatedAt:          n.UpdatedAt,
	}
}

func NewListNonconformitiesOutput(nonconformityPage *page.Page[*coredata.Nonconformity, coredata.NonconformityOrderField]) ListNonconformitiesOutput {
	nonconformities := make([]*Nonconformity, 0, len(nonconformityPage.Data))
	for _, v := range nonconformityPage.Data {
		nonconformities = append(nonconformities, NewNonconformity(v))
	}

	var nextCursor *page.CursorKey
	if len(nonconformityPage.Data) > 0 {
		cursorKey := nonconformityPage.Data[len(nonconformityPage.Data)-1].CursorKey(nonconformityPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListNonconformitiesOutput{
		NextCursor:      nextCursor,
		Nonconformities: nonconformities,
	}
}
