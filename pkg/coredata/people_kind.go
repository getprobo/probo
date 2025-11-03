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

import (
	"database/sql/driver"
	"fmt"
)

type (
	PeopleKind uint8
)

const (
	PeopleKindEmployee PeopleKind = iota
	PeopleKindContractor
	PeopleKindServiceAccount
)

func PeopleKinds() []PeopleKind {
	return []PeopleKind{
		PeopleKindEmployee,
		PeopleKindContractor,
		PeopleKindServiceAccount,
	}
}

func (ps PeopleKind) MarshalText() ([]byte, error) {
	return []byte(ps.String()), nil
}

func (ps *PeopleKind) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case PeopleKindEmployee.String():
		*ps = PeopleKindEmployee
	case PeopleKindContractor.String():
		*ps = PeopleKindContractor
	case PeopleKindServiceAccount.String():
		*ps = PeopleKindServiceAccount
	default:
		return fmt.Errorf("invalid PeopleKind value: %q", val)
	}

	return nil
}

func (ts PeopleKind) String() string {
	var val string

	switch ts {
	case PeopleKindEmployee:
		val = "EMPLOYEE"
	case PeopleKindContractor:
		val = "CONTRACTOR"
	case PeopleKindServiceAccount:
		val = "SERVICE_ACCOUNT"
	}

	return val
}

func (pk *PeopleKind) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for PeopleKind, expected string got %T", value)
	}

	return pk.UnmarshalText([]byte(val))
}

func (pk PeopleKind) Value() (driver.Value, error) {
	return pk.String(), nil
}
