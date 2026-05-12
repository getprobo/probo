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

type AccessEntryAuthMethod string

const (
	AccessEntryAuthMethodSSO            AccessEntryAuthMethod = "SSO"
	AccessEntryAuthMethodPassword       AccessEntryAuthMethod = "PASSWORD"
	AccessEntryAuthMethodAPIKey         AccessEntryAuthMethod = "API_KEY"
	AccessEntryAuthMethodServiceAccount AccessEntryAuthMethod = "SERVICE_ACCOUNT"
	AccessEntryAuthMethodUnknown        AccessEntryAuthMethod = "UNKNOWN"
)

func AccessEntryAuthMethods() []AccessEntryAuthMethod {
	return []AccessEntryAuthMethod{
		AccessEntryAuthMethodSSO,
		AccessEntryAuthMethodPassword,
		AccessEntryAuthMethodAPIKey,
		AccessEntryAuthMethodServiceAccount,
		AccessEntryAuthMethodUnknown,
	}
}

func (a AccessEntryAuthMethod) String() string {
	return string(a)
}

func (a *AccessEntryAuthMethod) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan AccessEntryAuthMethod: unsupported type %T", value)
	}

	switch str {
	case "SSO":
		*a = AccessEntryAuthMethodSSO
	case "PASSWORD":
		*a = AccessEntryAuthMethodPassword
	case "API_KEY":
		*a = AccessEntryAuthMethodAPIKey
	case "SERVICE_ACCOUNT":
		*a = AccessEntryAuthMethodServiceAccount
	case "UNKNOWN":
		*a = AccessEntryAuthMethodUnknown
	default:
		return fmt.Errorf("cannot parse AccessEntryAuthMethod: invalid value %q", str)
	}
	return nil
}

func (a AccessEntryAuthMethod) Value() (driver.Value, error) {
	return a.String(), nil
}
