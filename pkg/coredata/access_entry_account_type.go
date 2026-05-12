// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type AccessEntryAccountType string

const (
	AccessEntryAccountTypeUser           AccessEntryAccountType = "USER"
	AccessEntryAccountTypeServiceAccount AccessEntryAccountType = "SERVICE_ACCOUNT"
)

func AccessEntryAccountTypes() []AccessEntryAccountType {
	return []AccessEntryAccountType{
		AccessEntryAccountTypeUser,
		AccessEntryAccountTypeServiceAccount,
	}
}

func (a AccessEntryAccountType) String() string {
	return string(a)
}

func (a *AccessEntryAccountType) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan AccessEntryAccountType: unsupported type %T", value)
	}

	switch str {
	case "USER":
		*a = AccessEntryAccountTypeUser
	case "SERVICE_ACCOUNT":
		*a = AccessEntryAccountTypeServiceAccount
	default:
		return fmt.Errorf("cannot parse AccessEntryAccountType: invalid value %q", str)
	}
	return nil
}

func (a AccessEntryAccountType) Value() (driver.Value, error) {
	return a.String(), nil
}
