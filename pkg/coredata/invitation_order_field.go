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

import "fmt"

// InvitationOrderField defines the fields that can be used to order invitations
type InvitationOrderField string

// InvitationOrderField constants
const (
	InvitationOrderFieldFullName   InvitationOrderField = "FULL_NAME"
	InvitationOrderFieldEmail      InvitationOrderField = "EMAIL"
	InvitationOrderFieldRole       InvitationOrderField = "ROLE"
	InvitationOrderFieldCreatedAt  InvitationOrderField = "CREATED_AT"
	InvitationOrderFieldExpiresAt  InvitationOrderField = "EXPIRES_AT"
	InvitationOrderFieldAcceptedAt InvitationOrderField = "ACCEPTED_AT"
)

func (p InvitationOrderField) Column() string {
	switch p {
	case InvitationOrderFieldFullName:
		return "full_name"
	case InvitationOrderFieldEmail:
		return "email"
	case InvitationOrderFieldRole:
		return "role"
	case InvitationOrderFieldCreatedAt:
		return "created_at"
	case InvitationOrderFieldExpiresAt:
		return "expires_at"
	case InvitationOrderFieldAcceptedAt:
		return "accepted_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}

func (e InvitationOrderField) IsValid() bool {
	switch e {
	case InvitationOrderFieldFullName, InvitationOrderFieldEmail, InvitationOrderFieldRole, InvitationOrderFieldCreatedAt, InvitationOrderFieldExpiresAt, InvitationOrderFieldAcceptedAt:
		return true
	}
	return false
}

func (e InvitationOrderField) String() string {
	return string(e)
}

func (e *InvitationOrderField) UnmarshalText(text []byte) error {
	*e = InvitationOrderField(text)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid InvitationOrderField", string(text))
	}
	return nil
}

func (e InvitationOrderField) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}
