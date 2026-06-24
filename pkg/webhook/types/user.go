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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type (
	User struct {
		ID                       gid.GID                `json:"id"`
		OrganizationID           gid.GID                `json:"organizationId"`
		EmailAddress             mail.Addr              `json:"emailAddress"`
		FullName                 string                 `json:"fullName"`
		Kind                     *string                `json:"kind"`
		Source                   coredata.ProfileSource `json:"source"`
		AdditionalEmailAddresses mail.Addrs             `json:"additionalEmailAddresses"`
		Position                 *string                `json:"position"`
		ContractStartDate        *time.Time             `json:"contractStartDate"`
		ContractEndDate          *time.Time             `json:"contractEndDate"`
		CreatedAt                time.Time              `json:"createdAt"`
		UpdatedAt                time.Time              `json:"updatedAt"`
		Membership               *UserMembership        `json:"membership"`
	}

	UserMembership struct {
		ID    gid.GID                 `json:"id"`
		Role  coredata.MembershipRole `json:"role"`
		State coredata.ProfileState   `json:"state"`
	}
)

func NewUser(p *coredata.MembershipProfile, m *coredata.Membership) *User {
	u := &User{
		ID:                       p.ID,
		OrganizationID:           p.OrganizationID,
		EmailAddress:             p.EmailAddress,
		FullName:                 p.FullName,
		Kind:                     p.Kind,
		Source:                   p.Source,
		AdditionalEmailAddresses: p.AdditionalEmailAddresses,
		Position:                 p.Position,
		ContractStartDate:        p.ContractStartDate,
		ContractEndDate:          p.ContractEndDate,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
	}

	if m != nil {
		u.Membership = &UserMembership{
			ID:    m.ID,
			Role:  m.Role,
			State: p.State,
		}
	}

	return u
}
