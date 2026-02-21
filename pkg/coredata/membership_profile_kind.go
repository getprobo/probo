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
	MembershipProfileKind uint8
)

const (
	MembershipProfileKindEmployee MembershipProfileKind = iota
	MembershipProfileKindContractor
	MembershipProfileKindServiceAccount
)

func MembershipProfileKinds() []MembershipProfileKind {
	return []MembershipProfileKind{
		MembershipProfileKindEmployee,
		MembershipProfileKindContractor,
		MembershipProfileKindServiceAccount,
	}
}

func (mpk MembershipProfileKind) MarshalText() ([]byte, error) {
	return []byte(mpk.String()), nil
}

func (mpk *MembershipProfileKind) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case MembershipProfileKindEmployee.String():
		*mpk = MembershipProfileKindEmployee
	case MembershipProfileKindContractor.String():
		*mpk = MembershipProfileKindContractor
	case MembershipProfileKindServiceAccount.String():
		*mpk = MembershipProfileKindServiceAccount
	default:
		return fmt.Errorf("invalid MembershipProfileKind value: %q", val)
	}

	return nil
}

func (mpk MembershipProfileKind) String() string {
	var val string

	switch mpk {
	case MembershipProfileKindEmployee:
		val = "EMPLOYEE"
	case MembershipProfileKindContractor:
		val = "CONTRACTOR"
	case MembershipProfileKindServiceAccount:
		val = "SERVICE_ACCOUNT"
	}

	return val
}

func (mpk *MembershipProfileKind) Scan(value any) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		return fmt.Errorf("unsupported type for MembershipProfileKind: %T", value)
	}

	return mpk.UnmarshalText([]byte(val))
}

func (mpk MembershipProfileKind) Value() (driver.Value, error) {
	return mpk.String(), nil
}
