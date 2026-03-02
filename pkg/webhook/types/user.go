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
	"go.probo.inc/probo/pkg/mail"
)

type User struct {
	ID                       gid.GID                        `json:"id"`
	OrganizationID           gid.GID                        `json:"organizationId"`
	EmailAddress             mail.Addr                      `json:"emailAddress"`
	FullName                 string                         `json:"fullName"`
	Kind                     coredata.MembershipProfileKind `json:"kind"`
	Source                   coredata.ProfileSource         `json:"source"`
	State                    coredata.ProfileState          `json:"state"`
	AdditionalEmailAddresses mail.Addrs                     `json:"additionalEmailAddresses"`
	Position                 *string                        `json:"position"`
	ContractStartDate        *time.Time                     `json:"contractStartDate"`
	ContractEndDate          *time.Time                     `json:"contractEndDate"`
	CreatedAt                time.Time                      `json:"createdAt"`
	UpdatedAt                time.Time                      `json:"updatedAt"`
}

func NewUser(p *coredata.MembershipProfile) *User {
	return &User{
		ID:                       p.ID,
		OrganizationID:           p.OrganizationID,
		EmailAddress:             p.EmailAddress,
		FullName:                 p.FullName,
		Kind:                     p.Kind,
		Source:                   p.Source,
		State:                    p.State,
		AdditionalEmailAddresses: p.AdditionalEmailAddresses,
		Position:                 p.Position,
		ContractStartDate:        p.ContractStartDate,
		ContractEndDate:          p.ContractEndDate,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
	}
}
