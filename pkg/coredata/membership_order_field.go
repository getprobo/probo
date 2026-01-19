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

package coredata

type (
	MembershipOrderField string
)

const (
	MembershipOrderFieldOrganizationName MembershipOrderField = "ORGANIZATION_NAME"
	MembershipOrderFieldFullName         MembershipOrderField = "FULL_NAME"
	MembershipOrderFieldEmailAddress     MembershipOrderField = "EMAIL_ADDRESS"
	MembershipOrderFieldRole             MembershipOrderField = "ROLE"
	MembershipOrderFieldCreatedAt        MembershipOrderField = "CREATED_AT"
)

func (p MembershipOrderField) Column() string {
	switch p {
	case MembershipOrderFieldOrganizationName:
		return "organization_name"
	case MembershipOrderFieldFullName:
		return "full_name"
	case MembershipOrderFieldEmailAddress:
		return "email_address"
	case MembershipOrderFieldRole:
		return "role"
	case MembershipOrderFieldCreatedAt:
		return "created_at"
	}
	return string(p)
}

func (p MembershipOrderField) String() string {
	return string(p)
}

func (p MembershipOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *MembershipOrderField) UnmarshalText(text []byte) error {
	*p = MembershipOrderField(text)
	return nil
}
