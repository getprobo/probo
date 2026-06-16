// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/mail"
)

func NewProfile(p *coredata.MembershipProfile) *Profile {
	additionalEmailAddresses := p.AdditionalEmailAddresses
	if additionalEmailAddresses == nil {
		additionalEmailAddresses = mail.Addrs{}
	}

	return &Profile{
		ID:                       p.ID,
		OrganizationID:           p.OrganizationID,
		FullName:                 p.FullName,
		EmailAddress:             p.EmailAddress,
		AdditionalEmailAddresses: additionalEmailAddresses,
		Kind:                     p.Kind,
		Source:                   p.Source,
		State:                    p.State,
		Position:                 p.Position,
		ContractStartDate:        p.ContractStartDate,
		ContractEndDate:          p.ContractEndDate,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
	}
}
