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

// MembershipOrderField defines the fields that can be used to order memberships
type MembershipOrderField string

// MembershipOrderField constants
const (
	MembershipOrderFieldCreatedAt MembershipOrderField = "CREATED_AT"
	MembershipOrderFieldUpdatedAt MembershipOrderField = "UPDATED_AT"
)

func (p MembershipOrderField) Column() string {
	switch p {
	case MembershipOrderFieldCreatedAt:
		return "created_at"
	case MembershipOrderFieldUpdatedAt:
		return "updated_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}

// IsValid returns true if the order field is valid
func (e MembershipOrderField) IsValid() bool {
	switch e {
	case MembershipOrderFieldCreatedAt, MembershipOrderFieldUpdatedAt:
		return true
	}
	return false
}

func (e MembershipOrderField) String() string {
	return string(e)
}

func (e *MembershipOrderField) UnmarshalText(text []byte) error {
	*e = MembershipOrderField(text)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid MembershipOrderField", string(text))
	}
	return nil
}

func (e MembershipOrderField) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}
