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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type Obligation struct {
	ID                     gid.GID                  `json:"id"`
	OrganizationID         gid.GID                  `json:"organizationId"`
	Area                   *string                  `json:"area"`
	Source                 *string                  `json:"source"`
	Requirement            *string                  `json:"requirement"`
	ActionsToBeImplemented *string                  `json:"actionsToBeImplemented"`
	Regulator              *string                  `json:"regulator"`
	OwnerID                gid.GID                  `json:"ownerId"`
	LastReviewDate         *time.Time               `json:"lastReviewDate"`
	DueDate                *time.Time               `json:"dueDate"`
	Status                 coredata.ObligationStatus `json:"status"`
	Type                   coredata.ObligationType   `json:"type"`
	CreatedAt              time.Time                `json:"createdAt"`
	UpdatedAt              time.Time                `json:"updatedAt"`
}

func NewObligation(o *coredata.Obligation) *Obligation {
	return &Obligation{
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
}
